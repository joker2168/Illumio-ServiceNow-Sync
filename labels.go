package main

import (
	"encoding/json"

	"stash.ilabs.io/scm/~brian.pitta/illumioapi.git"
)

func checkAndCreateLabels(label illumioapi.Label, hostname string) (illumioapi.Label, bool, error) {
	config, pce := parseConfig()

	// CHECK IF LABEL EXISTS
	labelCheck, err := illumioapi.GetLabel(pce, label.Key, label.Value)
	if err != nil {
		return label, false, err
	}

	// IF LABEL DOESN'T EXIST, CREATE IT
	if len(labelCheck) == 0 {
		var l illumioapi.Label
		if config.Logging.LogOnly == false {
			newLabel, err := illumioapi.CreateLabel(pce, label)
			if err != nil {
				return label, false, err
			}
			json.Unmarshal([]byte(newLabel.RespBody), &l)
			return l, true, nil
		}
	}
	return labelCheck[0], false, nil
}
