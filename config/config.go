package config

import (
	"encoding/json"
	"log"
	"os"
)

type Log struct {
	Enabled bool `json:"enabled"`
	Debug   bool `json:"debug"`
}

type Actors struct {
	Known []string `json:"known"`
}

type Bluetooth struct {
	AdvertisementName        string `json:"advertisementName"`
	AdvertisementDelayMs     int    `json:"advertisementDelayMs"`
	ServiceID                string `json:"serviceId"`
	IndicateCharacteristicID string `json:"indicateCharacteristicId"`
	ConnectionPoolSize       int    `json:"connectionPoolSize"`
	ConnectionsLimit         int    `json:"connectionsLimit"`
	ConnectionLimitDelayMs   int    `json:"connectionLimitDelayMs"`
	DisconnectionDelayMs     int    `json:"disconnectionDelayMs"`
}

type Config struct {
	Bluetooth Bluetooth `json:"bluetooth"`
	Actors    Actors    `json:"actors"`
	Log       Log       `json:"log"`

	EventLoopDelayMs int `json:"eventLoopDelayMs"`
	RelayDebounceMs  int `json:"relayDebounceMs"`
	OperationDelayMs int `json:"operationDelayMs"`
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
