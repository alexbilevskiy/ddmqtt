//go:build windows

package main

import (
	"ddmqtt/config"
	"ddmqtt/hass"
	"ddmqtt/mqtt"
)

func main() {
	config.InitConfig("config.json")
	mqtt.InitMqtt()
	hass.StartReporting()
}
