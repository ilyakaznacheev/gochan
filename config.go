package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const (
	configFile = "./config.json"
)

// ConfigData contains app configuration data
type ConfigData struct {
	Database struct {
		User    string
		Name    string
		Pass    string
		Address string
		SSL     string
	}
	// map[string]string
	Redis struct {
		Address  string
		Password string
		DataBase int
	}
}

var config *ConfigData

func getConfig() *ConfigData {
	if config == nil {
		file, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal("config not found:", err)
		}

		config = &ConfigData{}

		err = json.Unmarshal(file, config)
		if err != nil {
			log.Fatal("config is corrupted:", err)
		}
	}
	return config
}
