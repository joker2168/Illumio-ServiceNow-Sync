package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"stash.ilabs.io/scm/~brian.pitta/illumioapi.git"
)

func main() {

	// GET CONFIG
	config, pce := parseConfig()

	// SET UP LOGGING FILE
	if len(config.Logging.LogDirectory) > 0 && config.Logging.LogDirectory[len(config.Logging.LogDirectory)-1:] != string(os.PathSeparator) {
		config.Logging.LogDirectory = config.Logging.LogDirectory + string(os.PathSeparator)
	}
	f, err := os.OpenFile(config.Logging.LogDirectory+"Illumio_ServiceNow_Sync_"+time.Now().Format("20060102_150405")+".log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	// LOG THE MODE
	log.Printf("INFO - Log only mode set to %t", config.Logging.LogOnly)
	if config.Logging.LogOnly == true {
		log.Printf("INFO - THIS MEANS ALL CHANGES LOGGED TO THE PCE DID NOT ACTUALLY HAPPEN. THEY WILL HAPPEN IF YOU RUN AGAIN WITH LOG ONLY SET TO FALSE.")
	}
	log.Printf("INFO - Create unmanaged workloads set to %t", config.UnmanagedWorkloads.Enable)

	// GET ALL EXISTING LABELS AHEAD OF TIME (SAVES API CALLS)
	labelsAPI, err := illumioapi.GetAllLabels(pce)
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
	wlAPI, err := illumioapi.GetAllWorkloads(pce)
	if err != nil {
		log.Fatal(err)
	}
	accountWorkloads := make(map[string]illumioapi.Workload)
	for _, w := range wlAPI {
		accountWorkloads[w.Href] = w
	}

	// GET DATA FROM SERVICENOW TABLE
	snURL := config.ServiceNow.TableURL + "?CSV&sysparm_fields=" + url.QueryEscape(config.ServiceNow.MatchField) + "," + url.QueryEscape(config.LabelMapping.App) +
		"," + url.QueryEscape(config.LabelMapping.Enviornment) + "," + url.QueryEscape(config.LabelMapping.Location) + "," + url.QueryEscape(config.LabelMapping.Role)

	if config.UnmanagedWorkloads.Enable == true && config.UnmanagedWorkloads.Table == "cmdb_ci_server_list" {
		snURL = snURL + ",ip_address,host_name"
	}
	data := snhttp(snURL)

	// SET THE TOTAL MATCH VARIABLE AND COUNTER
	counter := 0
	totalMatch := 0

	for _, line := range data {
		counter++
		lineMatch := 0

		updateLabelsArray := make([]illumioapi.Label, 0)
		// CHECK IF WORKLOAD EXISTS
		for _, wl := range accountWorkloads {

			// SET SOME WORKLOAD SPECIFIC VARIABLES
			updateRequired := false
			updateLabelsArray = nil
			wlLabels := make(map[string]string)

			// SWITCH THE MATCH FIELD FROM HOSTNAME BASED ON CONFIG
			illumioMatch := wl.Hostname
			if config.Illumio.MatchField == "name" {
				illumioMatch = wl.Name
			}

			// IF THE FIRST COL (MATCH) MATHCES THE ILLUMIO MATCH, TAKE ACTION
			if line[0] == illumioMatch {
				totalMatch++
				lineMatch++
				for _, l := range wl.Labels {
					wlLabels[accountLabelKeys[l.Href]] = accountLabelValues[l.Href]
				}
				// CHECK EACH LABEL TYPE TO SEE IF IT NEEDS TO BE UPDATED
				labelKeys := []string{"app", "env", "loc", "role"}
				configFields := []string{config.LabelMapping.App, config.LabelMapping.Enviornment, config.LabelMapping.Location, config.LabelMapping.Role}

				for i := 0; i <= 3; i++ {
					// ONLY WORK ON COLUMNS THAT ARE NOT "csvPlaceHolderIllumio" COLUMNS (SET IN CONFIG PARSING) AND LABELS DON'T MATCH
					if configFields[i] != "csvPlaceHolderIllumio" && wlLabels[labelKeys[i]] != line[i+1] {
						log.Printf("INFO - %s - %s label updated from %s to %s", wl.Hostname, labelKeys[i], wlLabels[labelKeys[i]], line[i+1])
						updateRequired = true
						// IF THE NEW VALUE (FROM SN) IS BLANK, WE DON'T APPEND TO THE UPDATE ARRAY
						if line[i+1] != "" {
							updateLabelsArray = append(updateLabelsArray, illumioapi.Label{Key: labelKeys[i], Value: line[i+1]})
						}
						// ADD EXISTING LABEL IF IT EXISTS
					} else if line[i+1] != "" {
						updateLabelsArray = append(updateLabelsArray, illumioapi.Label{Key: labelKeys[i], Value: wlLabels[labelKeys[i]]})
					}
				}

				// UPDATE THE WORKLOAD IF ANYTHING NEEDS TO CHANGE
				if updateRequired == true {

					// GET LABEL HREF AND CREATE IF NECESSARY
					ulHrefs := []*illumioapi.Label{}
					for _, ul := range updateLabelsArray {
						labelCheck, createdBool, err := checkAndCreateLabels(ul, wl.Hostname)
						if err != nil {
							log.Printf("ERROR - %s - %s", wl.Hostname, err)
						}
						if createdBool == true {
							log.Printf("INFO - CREATED LABEL %s (%s)", ul.Value, ul.Key)
						}
						ulHrefs = append(ulHrefs, &illumioapi.Label{Href: labelCheck.Href})
					}

					// UPDATE THE WORKLOAD
					payload := illumioapi.Workload{Href: wl.Href, Labels: ulHrefs}

					if config.Logging.LogOnly == false {
						_, err := illumioapi.UpdateWorkload(pce, payload)
						if err != nil {
							log.Printf("ERROR - %s - UpdateWorkLoad - %s - Updates did not get pushed to PCE", wl.Hostname, err)
						}
					}

				} else {
					log.Printf("INFO - %s - No label updates required", wl.Hostname)
				}
			}

		}
		// IF THERE WERE NO MATCHES AND IT'S NOT THE HEADER FILE, CREATE THE UNMANAGED WORKLOAD
		if lineMatch == 0 && counter != 1 && config.UnmanagedWorkloads.Enable == true {
			if len(line[6]) > 0 && len(line[5]) > 0 {

				labelKeys := []string{"app", "env", "loc", "role"}
				configFields := []string{config.LabelMapping.App, config.LabelMapping.Enviornment, config.LabelMapping.Location, config.LabelMapping.Role}
				var labelArray []*illumioapi.Label
				for i := 0; i <= 3; i++ {
					// ONLY WORK ON COLUMNS THAT ARE NOT "csvPlaceHolderIllumio" COLUMNS (SET IN CONFIG PARSING) AND LABELS DON'T MATCH
					if configFields[i] != "csvPlaceHolderIllumio" && len(line[i+1]) > 0 {
						l, create, err := checkAndCreateLabels(illumioapi.Label{Key: labelKeys[i], Value: line[i+1]}, line[6])
						if err != nil {
							log.Printf("ERROR - %s", err)
						}
						if create == true {
							log.Printf("INFO - CREATED LABEL %s (%s)", line[i+1], labelKeys[i])
						}
						labelArray = append(labelArray, &illumioapi.Label{Href: l.Href})
					}
				}

				intfaces := []*illumioapi.Interface{&illumioapi.Interface{Name: "eth0", Address: line[5]}}
				umwl := illumioapi.Workload{
					Name:       line[6],
					Hostname:   line[6],
					Interfaces: intfaces,
					Labels:     labelArray}
				_, err := illumioapi.CreateWorkload(pce, umwl)
				if err != nil {
					log.Printf("ERROR - Could not create workload %s with IP address %s - %s", line[6], line[5], err)
				}
				if err == nil {
					log.Printf("INFO - Created unmanaged workload for hostname %s", line[6])
				}
			} else {
				log.Printf("WARNING - Could not create unmanaged workload for hostname %s with IP Address %s - not enough information.", line[6], line[5])
			}
		}

	}
	log.Printf("INFO - Identified %d servers in CMDB - %d in the PCE and %d not in PCE", len(data)-1, totalMatch, len(data)-1-totalMatch)
}
