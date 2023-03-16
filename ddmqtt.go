//go:build windows

package main

import (
	"ddmqtt/ddmrpc"
	"ddmqtt/structs"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var c mqtt.Client
var cfg structs.Config

func main() {
	configFile := "config.json"

	cfgRaw, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cannot open config: %s", err.Error())
	}
	err = json.Unmarshal(cfgRaw, &cfg)
	if err != nil {
		log.Fatalf("cannot parse config file: %s", err.Error())
	}

	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(cfg.BrokerAddr).SetClientID(cfg.MqttClientId)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect: %s", token.Error())
	}

	//if token := c.Subscribe(fmt.Sprintf("%s/#", cfg.HassDiscoveryPrefix), 0, nil); token.Wait() && token.Error() != nil {
	//	log.Fatalf("failed to subscribe: %s", token.Error())
	//}

	discovery()
	select {}
}

func discovery() {
	attrs, err := ddmrpc.GetAssetAttributes()
	if err != nil {

		log.Fatalf("failed to read monitor info: %s", err.Error())
	}

	monitor := structs.DiscoveryDevice{
		Identifiers:  attrs.ServiceTag,
		Manufacturer: "Dell",
		Model:        attrs.ModelCode,
		Name:         attrs.Model,
	}
	baseTopic := fmt.Sprintf("%s/sensor/%s/active_hours", cfg.HassDiscoveryPrefix, monitor.Identifiers)

	sensor := structs.DiscoverySensor{
		Name: "Active hours",
		//DeviceClass:       nil,
		StateTopic:        fmt.Sprintf("%s/state", baseTopic),
		AvailabilityTopic: fmt.Sprintf("%s/available", baseTopic),
		//Availability: structs.SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		UniqueId: fmt.Sprintf("%s_hours", attrs.ServiceTag),
		Device:   monitor,
	}
	discoveryTopic := fmt.Sprintf("%s/config", baseTopic)
	js, _ := json.Marshal(sensor)
	log.Printf("publishing discovery to: %s / %s", discoveryTopic, string(js))

	pubToken := c.Publish(discoveryTopic, 0, false, js)
	if pubToken.Error() != nil {
		log.Fatalf("failed to publish discovery: %s", pubToken.Error())
	}

	log.Printf("publishing availability to: %s / %s", sensor.AvailabilityTopic, "online")

	pubOnlineToken := c.Publish(sensor.AvailabilityTopic, 0, false, "online")
	if pubOnlineToken.Error() != nil {
		log.Fatalf("failed to publish online state: %s", pubOnlineToken.Error())
	}

	log.Printf("publishing state to: %s / %s", sensor.StateTopic, fmt.Sprintf("%d", attrs.ActiveHours))

	pubState := c.Publish(sensor.StateTopic, 0, false, fmt.Sprintf("%d", attrs.ActiveHours))
	if pubState.Error() != nil {
		log.Fatalf("failed to publish state data: %s", pubState.Error())
	}

	if token := c.Subscribe(fmt.Sprintf("%s/#", baseTopic), 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe: %s", token.Error())
	}

}
