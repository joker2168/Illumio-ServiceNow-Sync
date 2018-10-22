package main

import (
	"crypto/tls"
	"encoding/csv"
	"log"
	"net/http"
)

func snhttp(url string) [][]string {

	// GET CONFIG
	config, _ := parseConfig()

	// CREATE HTTP CLIENT
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// SET BASIC AUTH
	req.SetBasicAuth(config.ServiceNow.User, config.ServiceNow.Password)

	// MAKE HTTP REQUEST
	resp, err := client.Do(req)

	// LOG SERVICE NOW HTTP REQUEST
	if config.Logging.logLevel == true {
		log.Printf("DEBUG - ServiceNow API Response Status Code: %d \r\n", resp.StatusCode)
	}
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
	if config.Logging.logLevel == true {
		log.Printf("DEBUG - ServiceNowAPI Response CSV Data: %v \r\n", data)
	}

	return data
}
