//go:build windows

package main

import (
	"ddmqtt/config"
	"ddmqtt/hass"
	"ddmqtt/mqtt"
)

func main() {
	config.InitConfig("config.json")
	monitor := hass.Prepare()
	mqtt.InitMqtt(monitor.Identifiers)
	hass.StartReporting(monitor)
}
