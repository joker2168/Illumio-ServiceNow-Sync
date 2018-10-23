package main

import (
	"log"

	illumioapi "stash.ilabs.io/scm/~brian.pitta/illumioapi.git"
)

func updateWorkload(labels []illumioapi.Label, workload illumioapi.Workload) {

	// GET CONFIG
	config, pce := parseConfig()

	// GET LABEL HREF AND CREATE IF NECESSARY
	ulHrefs := []*illumioapi.Label{}
	for _, ul := range labels {
		labelCheck := checkAndCreateLabels(ul, workload.Hostname)
		ulHrefs = append(ulHrefs, &illumioapi.Label{Href: labelCheck.Href})
	}

	// UPDATE THE WORKLOAD
	payload := illumioapi.Workload{Href: workload.Href, Labels: ulHrefs}

	if config.Logging.LogOnly == false {
		updateWlAPI, err := illumioapi.UpdateWorkload(pce, payload)
		if config.Logging.verbose == true {
			log.Printf("DEBUG - Update WL API HTTP Request: %s %v \r\n", updateWlAPI.Request.Method, updateWlAPI.Request.URL)
			log.Printf("DEBUG - Update WL API HTTP Reqest Header: %v \r\n", updateWlAPI.Request.Header)
			log.Printf("DEBUG - Update WL API for %s Response Status Code: %d \r\n", workload.Hostname, updateWlAPI.StatusCode)
			log.Printf("DEBUG - Update WL API for %s Response Body: %s \r\n", workload.Hostname, updateWlAPI.RespBody)
		}
		if err != nil {
			log.Printf("ERROR - %s - UpdateWorkLoad - %s - Updates did not get pushed to PCE", workload.Hostname, err)
		}
	}
}

func createUnmanagedWorkload(interfaceList, ipAddressList []string, app, env, loc, role, hostname string) error {

	// GET CONFIG
	config, pce := parseConfig()

	labels := []string{app, env, loc, role}
	labelKeys := []string{"app", "env", "loc", "role"}
	configFields := []string{config.LabelMapping.App, config.LabelMapping.Enviornment, config.LabelMapping.Location, config.LabelMapping.Role}
	var labelArray []*illumioapi.Label

	for i := 0; i <= 3; i++ {
		// ONLY WORK ON COLUMNS THAT ARE NOT "csvPlaceHolderIllumio" COLUMNS (SET IN CONFIG PARSING) AND LABELS DON'T MATCH
		if configFields[i] != "csvPlaceHolderIllumio" && len(labels[i]) > 0 {
			l := checkAndCreateLabels(illumioapi.Label{Key: labelKeys[i], Value: labels[i]}, hostname)
			labelArray = append(labelArray, &illumioapi.Label{Href: l.Href})
		}
	}

	var networkInterfaces []*illumioapi.Interface
	counter := 0
	for _, networkinterface := range interfaceList {
		networkInterfaces = append(networkInterfaces, &illumioapi.Interface{Name: networkinterface, Address: ipAddressList[counter]})
		counter++
	}
	umwl := illumioapi.Workload{
		Name:       hostname,
		Hostname:   hostname,
		Interfaces: networkInterfaces,
		Labels:     labelArray}
	_, err := illumioapi.CreateWorkload(pce, umwl)
	if err != nil {
		log.Printf("ERROR - Could not create workload %s - %s", hostname, err)
		return err
	}
	if err == nil {
		log.Printf("INFO - Created unmanaged workload for hostname %s", hostname)
	}
	return nil
}
