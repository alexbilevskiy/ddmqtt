package config

import (
	"encoding/json"
	"log"
	"os"
)

func InitConfig(configFile string) *Config {
	var cfg Config

	cfgRaw, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cannot open config: %s", err.Error())
	}
	err = json.Unmarshal(cfgRaw, &cfg)
	if err != nil {
		log.Fatalf("cannot parse config file: %s", err.Error())
	}

	return &cfg
}
