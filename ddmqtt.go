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

	//if token := c.Subscribe(fmt.Sprintf("%s/#", cfg.HassDiscoveryPrefix), 0, nil); token.Wait() && token.Error() != nil {
	//	log.Fatalf("failed to subscribe: %s", token.Error())
	//}

	hass.StartReporting()
}
