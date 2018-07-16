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

	return data
}
