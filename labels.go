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
	labelCheck, err := illumioapi.GetLabel(pce, label.Key, label.Value)
	if err != nil {
		log.Printf("ERROR - Cannot check if %s (%s) exists - %s", label.Value, label.Key, err)
	}

	// IF LABEL DOESN'T EXIST, CREATE IT
	if labelCheck.Key == "" {
		if config.Logging.LogOnly == false {
			newLabel, err := illumioapi.CreateLabel(pce, label)
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
