package config

import (
	"ddmqtt/structs"
	"encoding/json"
	"log"
	"os"
)

var CFG structs.Config

func InitConfig(configFile string) {
	cfgRaw, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cannot open config: %s", err.Error())
	}
	err = json.Unmarshal(cfgRaw, &CFG)
	if err != nil {
		log.Fatalf("cannot parse config file: %s", err.Error())
	}
}
