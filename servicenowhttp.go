package main

import (
	"crypto/tls"
	"encoding/csv"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	if config.Logging.verbose == true {
		log.Printf("DEBUG - Making ServiceNow API call ...\r\n")
	}
	resp, err := client.Do(req)

	// LOG SERVICE NOW HTTP REQUEST
	if config.Logging.verbose == true {
		log.Printf("DEBUG - ServiceNow API HTTP Request Made: %s %v \r\n", resp.Request.Method, resp.Request.URL)
		log.Printf("DEBUG - ServiceNow API Response Status Code: %d \r\n", resp.StatusCode)

	}
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// READ IN CSV DATA
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	reader := csv.NewReader(strings.NewReader(bodyString))
	data, err := reader.ReadAll()
	if config.Logging.verbose == true {
		log.Printf("DEBUG - ServiceNowAPI Response CSV Data:\r\n %v \r\n", strings.Replace(bodyString, "\n", "\r\n", -1))
	}
	if err != nil {
		log.Fatalf("ERROR - %s", err)
	}
	return data
}
