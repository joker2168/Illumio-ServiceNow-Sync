package main

import (
	"encoding/json"
	"log"

	"github.com/brian1917/illumioapi"
)

func checkAndCreateLabels(label illumioapi.Label, hostname string) illumioapi.Label {
	config, pce := parseConfig()

	var l illumioapi.Label

	// CHECK IF LABEL EXISTS
	labelCheck, apiResp, err := illumioapi.GetLabelbyKeyValue(pce, label.Key, label.Value)
	if config.Logging.verbose == true {
		log.Printf("DEBUG - Get Label API HTTP Request: %s %v \r\n", apiResp.Request.Method, apiResp.Request.URL)
		log.Printf("DEBUG - Get Label API HTTP Reqest Header: %v \r\n", apiResp.Request.Header)
		log.Printf("DEBUG - Get Label API for %s (%s) Response Status Code: %d \r\n", label.Value, label.Key, apiResp.StatusCode)
		log.Printf("DEBUG - Get Label API for %s (%s) Response Body: %s \r\n", label.Value, label.Key, apiResp.RespBody)
	}
	if err != nil {
		log.Printf("ERROR - Cannot check if %s (%s) exists - %s", label.Value, label.Key, err)
	}

	// IF LABEL DOESN'T EXIST, CREATE IT
	if labelCheck.Key == "" {
		if config.Logging.LogOnly == false {
			_, newLabel, err := illumioapi.CreateLabel(pce, label)
			if config.Logging.verbose == true {
				log.Printf("DEBUG - Exact label does not exist for %s (%s). Creating new label... \r\n", label.Value, label.Key)
				log.Printf("DEBUG - Create Label API HTTP Request: %s %v \r\n", newLabel.Request.Method, newLabel.Request.URL)
				log.Printf("DEBUG - Create Label API HTTP Reqest Header: %v \r\n", newLabel.Request.Header)
				log.Printf("DEBUG - Create Label API for %s (%s) Response Status Code: %d \r\n", label.Value, label.Key, newLabel.StatusCode)
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
