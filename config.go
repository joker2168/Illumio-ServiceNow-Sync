package main

import (
	"flag"
	"io/ioutil"
	"log"

	"stash.ilabs.io/scm/~brian.pitta/illumioapi.git"

	"github.com/BurntSushi/toml"
)

type config struct {
	Illumio            illumio      `toml:"illumio"`
	ServiceNow         serviceNow   `toml:"serviceNow"`
	LabelMapping       labelMapping `toml:"labelMapping"`
	Logging            logging      `toml:"logging"`
	UnmanagedWorkloads unmanagedWLs `toml:"unmanagedWorkloads"`
}

type illumio struct {
	FQDN       string `toml:"fqdn"`
	Port       int    `toml:"port"`
	Org        int    `toml:"org"`
	User       string `toml:"user"`
	Key        string `toml:"key"`
	MatchField string `toml:"match_field"`
}

type serviceNow struct {
	TableURL   string `toml:"table_url"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	MatchField string `toml:"match_field"`
}

type labelMapping struct {
	App         string `toml:"app"`
	Enviornment string `toml:"enviornment"`
	Location    string `toml:"location"`
	Role        string `toml:"role"`
}

type logging struct {
	LogOnly      bool   `toml:"log_only"`
	LogDirectory string `toml:"log_directory"`
}

type unmanagedWLs struct {
	Enable bool   `toml:"enable"`
	Table  string `toml:"table"`
}

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "config.json", "location of JSON configuration file")
}

func parseConfig() (config, illumioapi.PCE) {
	var config config

	flag.Parse()

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	_, err = toml.Decode(string(data), &config)
	if err != nil {
		log.Fatal(err)
	}

	// IF A FIELD IS LEFT BLANK WE WANT TO PUT A PLACEHOLDER
	fields := []*string{&config.LabelMapping.App, &config.LabelMapping.Enviornment, &config.LabelMapping.Location, &config.LabelMapping.Role}
	for _, field := range fields {
		if *field == "" {
			*field = "csvPlaceHolderIllumio"
		}

	}

	pce := illumioapi.PCE{
		FQDN: config.Illumio.FQDN,
		Port: config.Illumio.Port,
		Org:  config.Illumio.Org,
		User: config.Illumio.User,
		Key:  config.Illumio.Key}

	return config, pce
}
