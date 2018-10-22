package main

import (
	"encoding/json"
	"log"

	"stash.ilabs.io/scm/~brian.pitta/illumioapi.git"
)

func checkAndCreateLabels(label illumioapi.Label, hostname string) illumioapi.Label {
	config, pce := parseConfig()

	var l illumioapi.Label

	// CHECK IF LABEL EXISTS
	labelCheck, apiResp, err := illumioapi.GetLabel(pce, label.Key, label.Value)
	if config.Logging.logLevel == true {
		log.Printf("DEBUG - Get Label API for %s (%s) Response Status Code: %d \r\n", label.Value, label.Key, apiResp.StatusCode)
		log.Printf("DEBIG - Get Label API for %s (%s) Response Headers: %s \r\n", label.Value, label.Key, apiResp.Header)
		log.Printf("DEBUG - Get Label API for %s (%s) Response Body: %s \r\n", label.Value, label.Key, apiResp.RespBody)
	}
	if err != nil {
		log.Printf("ERROR - Cannot check if %s (%s) exists - %s", label.Value, label.Key, err)
	}

	// IF LABEL DOESN'T EXIST, CREATE IT
	if labelCheck.Key == "" {
		if config.Logging.LogOnly == false {
			newLabel, err := illumioapi.CreateLabel(pce, label)
			if config.Logging.logLevel == true {
				log.Printf("DEBUG - Create Label API for %s (%s) Response Status Code: %d \r\n", label.Value, label.Key, newLabel.StatusCode)
				log.Printf("DEBIG - Create Label API for %s (%s) Response Headers: %s \r\n", label.Value, label.Key, newLabel.Header)
				log.Printf("DEBUG - Create Label API for %s (%s) Response Body: %s \r\n", label.Value, label.Key, newLabel.RespBody)
			}
			if err != nil {
				log.Printf("ERROR - %s - %s", hostname, err)
				return l
			}
			json.Unmarshal([]byte(newLabel.RespBody), &l)
			log.Printf("INFO - CREATED LABEL %s (%s)", label.Value, label.Key)
			return l
		}
	}
	return labelCheck
}
