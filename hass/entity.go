package hass

import (
	"ddmqtt/config"
	"ddmqtt/device"
	"ddmqtt/mqtt"
	"ddmqtt/structs"
	"encoding/json"
	"fmt"
	"log"
)

var needDiscovery = true

func ReportState() {
	monitor := device.InitMontor()
	SensorActiveHours(monitor)
}

func doDiscovery(sensor structs.DiscoverySensor, baseTopic string) {
	discoveryTopic := fmt.Sprintf("%s/config", baseTopic)

	js, _ := json.Marshal(sensor)
	log.Printf("publishing discovery to: %s / %s", discoveryTopic, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, false, js)
	if pubToken.Error() != nil {
		log.Fatalf("failed to publish discovery: %s", pubToken.Error())
	}
}

func SensorActiveHours(monitor structs.DiscoveryDevice) {
	objectId := fmt.Sprintf("%s_active_hours", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/sensor/%s", config.CFG.HassDiscoveryPrefix, objectId)
	sensor := structs.DiscoverySensor{
		Name:         "Active hours",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: structs.SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
	}

	if needDiscovery {
		doDiscovery(sensor, baseTopic)
		needDiscovery = false
	}

	activeHours, err := device.GetMonitorActiveHours()
	if err != nil {
		log.Printf("cannot get active hours: %s", err.Error())
		doAvailability(sensor.Availability.Topic, false)
		return
	}
	doAvailability(sensor.Availability.Topic, true)

	doState(sensor.StateTopic, fmt.Sprintf("%d", activeHours))

	if token := mqtt.C.Subscribe(fmt.Sprintf("%s/#", baseTopic), 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe: %s", token.Error())
	}
}

func doAvailability(availabilityTopic string, available bool) {
	availabilityStatus := "offline"
	if available {
		availabilityStatus = "online"
	}
	log.Printf("publishing availability to: %s / %s", availabilityTopic, availabilityStatus)

	pubOnlineToken := mqtt.C.Publish(availabilityTopic, 0, false, availabilityStatus)
	if pubOnlineToken.Error() != nil {
		log.Fatalf("failed to publish online state: %s", pubOnlineToken.Error())
	}
}

func doState(stateTopic string, state string) {
	log.Printf("publishing state to: %s / %s", stateTopic, state)

	pubState := mqtt.C.Publish(stateTopic, 0, false, state)
	if pubState.Error() != nil {
		log.Fatalf("failed to publish state data: %s", pubState.Error())
	}
}
