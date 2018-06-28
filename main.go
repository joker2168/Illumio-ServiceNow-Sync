package main

import (
	"crypto/tls"
	"encoding/csv"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"stash.ilabs.io/scm/~brian.pitta/illumioapi.git"
)

func main() {

	// GET CONFIG
	config := parseConfig()

	// SET UP LOGGING FILE
	f, err := os.OpenFile("Illumio_ServiceNow_Sync_"+time.Now().Format("20060102_150405")+".log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	// LOG THE MODE
	log.Printf("INFO - Log only mode set to %t", config.LoggingOnly)
	if config.LoggingOnly == true {
		log.Printf("INFO - THIS MEANS ALL CHANGES LOGGED TO THE PCE DID NOT ACTUALLY HAPPEN. THEY WILL HAPPEN IF YOU RUN AGAIN WITH LOG ONLY SET TO FALSE.")
	}

	// GET ALL EXISTING LABELS AHEAD OF TIME (SAVES API CALLS)
	labelsAPI, err := illumioapi.GetAllLabels(config.IllumioFQDN, config.IllumioPort, config.IllumioOrg, config.IllumioUser, config.IllumioKey)
	if err != nil {
		log.Fatal(err)
	}

	accountLabelKeys := make(map[string]string)
	accountLabelValues := make(map[string]string)

	for _, l := range labelsAPI {
		accountLabelKeys[l.Href] = l.Key
		accountLabelValues[l.Href] = l.Value
	}

	// GET ALL WORKLOADS AHEAD OF TIME (SAVES API CALLS)
	wlAPI, err := illumioapi.GetAllWorkloads(config.IllumioFQDN, config.IllumioPort, config.IllumioOrg, config.IllumioUser, config.IllumioKey)
	if err != nil {
		log.Fatal(err)
	}
	accountWorkloads := make(map[string]illumioapi.Workload)
	for _, w := range wlAPI {
		accountWorkloads[w.Href] = w
	}

	// CREATE HTTP CLIENT FOR SERVICENOW REQUEST
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	snURL := config.ServiceNowURL + "?CSV&sysparm_fields=" + url.QueryEscape(config.ServiceNowMatchField) + "," + url.QueryEscape(config.AppField) + "," + url.QueryEscape(config.EnvField) + "," + url.QueryEscape(config.LocField) + "," + url.QueryEscape(config.RoleField)
	req, err := http.NewRequest("GET", snURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	// SET BASIC AUTH
	req.SetBasicAuth(config.ServiceNowUser, config.ServiceNowPwd)

	// MAKE HTTP REQUEST
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// READ IN CSV DATA
	reader := csv.NewReader(resp.Body)
	data, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("ERROR - %s", err)
	}

	notInPCE := 0
	counter := 0
	for _, line := range data {

		// CHECK IF WORKLOAD EXISTS
		for _, wl := range accountWorkloads {
			counter++
			updateRequired := false
			wlLabels := make(map[string]string)
			illumioMatch := wl.Hostname
			if config.IllumioMatchField == "name" {
				illumioMatch = wl.Name
			}
			if line[0] == illumioMatch {
				for _, l := range wl.Labels {
					wlLabels[accountLabelKeys[l.Href]] = accountLabelValues[l.Href]
				}
				// CHECK EACH LABEL TYPE TO SEE IF IT NEEDS TO BE UPDATED
				labelKeys := []string{"app", "env", "loc", "role"}
				updateLabelsArray := []illumioapi.Label{}
				for i := 0; i <= 3; i++ {
					// CHECK IF THE WORKLOAD LABEL MATCHES THE CSV FIELD
					if wlLabels[labelKeys[i]] != line[i+1] {
						log.Printf("INFO - %s - %s label updated from %s to %s", wl.Hostname, labelKeys[i], wlLabels[labelKeys[i]], line[i+1])
						updateRequired = true
						updateLabelsArray = append(updateLabelsArray, illumioapi.Label{Key: labelKeys[i], Value: line[i+1]})
					} else {
						updateLabelsArray = append(updateLabelsArray, illumioapi.Label{Key: labelKeys[i], Value: wlLabels[labelKeys[i]]})
					}
				}

				// UPDATE THE WORKLOAD IF ANYTHING NEEDS TO CHANGE
				if updateRequired == true {

					// MAKE SURE THE LABEL EXISTS
					for _, ul := range updateLabelsArray {
						labelCheck, err := illumioapi.GetLabel(config.IllumioFQDN, config.IllumioPort, config.IllumioOrg, config.IllumioUser, config.IllumioKey, ul.Key, ul.Value)
						if err != nil {
							log.Printf("ERROR - %s - %s", wl.Hostname, err)
						}
						// IF LABEL DOESN'T EXIST, CREATE IT
						if len(labelCheck) == 0 && config.LoggingOnly == false {
							_, err := illumioapi.CreateLabel(config.IllumioFQDN, config.IllumioPort, config.IllumioOrg, config.IllumioUser, config.IllumioKey, ul)
							if err != nil {
								log.Printf("ERROR - %s - %s", wl.Hostname, err)
							}
						}
					}

					// UPDATE THE WORKLOAD
					workloadUpdates := []*illumioapi.Label{}
					for _, ul := range updateLabelsArray {
						label, err := illumioapi.GetLabel(config.IllumioFQDN, config.IllumioPort, config.IllumioOrg, config.IllumioUser, config.IllumioKey, ul.Key, ul.Value)
						if err != nil {
							log.Printf("ERROR - %s - %s", wl.Hostname, err)
						}
						workloadUpdates = append(workloadUpdates, &illumioapi.Label{Href: label[0].Href})
					}
					payload := illumioapi.Workload{Href: wl.Href, Labels: workloadUpdates}

					if config.LoggingOnly == false {
						_, err := illumioapi.UpdateWorkload(config.IllumioFQDN, config.IllumioPort, config.IllumioUser, config.IllumioKey, payload)
						if err != nil {
							log.Printf("ERROR - %s - %s", wl.Hostname, err)
						}
					}
				} else {
					log.Printf("INFO - %s - No label updates required", wl.Hostname)
				}
			} else {
				notInPCE++
			}

		}
	}
	log.Printf("INFO - Processed %d servers; %d not in PCE as workloads", counter-1, notInPCE-1)
}
