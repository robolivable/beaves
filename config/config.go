package config

import (
	"encoding/json"
	"log"
	"os"
)

type Actors struct {
	Known []string `json:"known"`
}

type Bluetooth struct {
	AdvertisementName        string `json:"advertisementName"`
	ServiceID                string `json:"serviceId"`
	IndicateCharacteristicID string `json:"indicateCharacteristicId"`
	ConnectionPoolSize       int    `json:"connectionPoolSize"`
}

type Config struct {
	Bluetooth Bluetooth `json:"bluetooth"`
	Actors    Actors    `json:"actors"`
}

var RuntimeConfig Config

const ConfigFile = "config.json"

func init() {
	file, err := os.Open(ConfigFile)
	if err != nil {
		log.Fatalf("app requires a %s file", ConfigFile)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&RuntimeConfig); err != nil {
		log.Fatalf("error decoding config file: %v", err.Error())
	}
}
