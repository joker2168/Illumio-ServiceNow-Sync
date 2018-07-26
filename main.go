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

	// SET UP LOGGING
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

	// GET ALL EXISTING LABELS
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

	// GET ALL EXISTING WORKLOADS
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
	newUnmanagedWLs := 0

	// ITERATE THROUGH EACH LINE OF THE CSV
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

				// ITERATE THROUGH EACH LABEL TYPE
				for i := 0; i <= 3; i++ {

					// CANNOT BE "csvPlaceHolderIllumio" (SKIPPING THAT COL) AND THE LABELS DON'T MATCH
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
					updateWorkload(updateLabelsArray, wl)
				} else {
					log.Printf("INFO - %s - No label updates required", wl.Hostname)
				}
			}

		}

		// IF THERE WERE NO MATCHES AND IT'S NOT THE HEADER FILE, CREATE THE UNMANAGED WORKLOAD
		if lineMatch == 0 && counter != 1 && config.UnmanagedWorkloads.Enable == true {
			interfaceList := []string{"eth0"}
			ipAddressList := []string{line[5]}
			if len(ipAddressList[0]) == 0 || len(line[0]) == 0 {
				log.Printf("WARNING - Not enough information to create unmanaged workload for hostname %s", line[0])
			} else {
				err := createUnmanagedWorkload(interfaceList, ipAddressList, line[1], line[2], line[3], line[4], line[0])
				if err == nil {
					newUnmanagedWLs++
				}

			}
		}

	}
	// SUMMARIZE ACTIONS FOR LOG
	log.Printf("INFO - %d total servers in CMDB", len(data)-1)
	log.Printf("INFO - %d in the PCE", totalMatch)
	if config.UnmanagedWorkloads.Enable == true {
		log.Printf("INFO - %d new unmanaged workloads created", newUnmanagedWLs)
		log.Printf("INFO - %d servers with not enough info for unmanaged workload.", len(data)-1-totalMatch-newUnmanagedWLs)
	}
}
