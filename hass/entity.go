package hass

import (
	"ddmqtt/config"
	"ddmqtt/device"
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	"log"
)

var needDiscovery = make(map[string]bool, 0)
var monitor Device

func ReportState() {
	if monitor == (Device{}) {
		attrs, err := device.GetAssetAttributes()
		if err != nil {
			log.Fatalf("failed to read monitor: %s", err.Error())
		}
		monitor = Device{
			Identifiers:  attrs.ServiceTag,
			Manufacturer: "Dell",
			Model:        attrs.ModelCode,
			Name:         attrs.Model,
		}
	}

	SensorActiveHours(monitor)
	NumberBrightness(monitor)
}

func doDiscovery(entity Entity, baseTopic string) {
	var objectId string
	switch entity.GetType() {
	case TYPE_SENSOR:
		objectId = entity.(Sensor).ObjectId
	case TYPE_NUMBER:
		objectId = entity.(Number).ObjectId
	}
	if _, ok := needDiscovery[objectId]; ok {
		return
	}
	needDiscovery[objectId] = false

	discoveryTopic := fmt.Sprintf("%s/config", baseTopic)

	js, _ := json.Marshal(entity)
	log.Printf("publishing discovery to: %s / %s", discoveryTopic, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, false, js)
	if pubToken.Error() != nil {
		log.Fatalf("failed to publish discovery: %s", pubToken.Error())
	}
}

func SensorActiveHours(monitor Device) {
	objectId := fmt.Sprintf("%s_active_hours", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/sensor/%s", config.CFG.HassDiscoveryPrefix, objectId)
	hours := Sensor{
		Name:         "Active hours",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
	}

	doDiscovery(hours, baseTopic)

	activeHours, err := device.GetMonitorActiveHours()
	if err != nil {
		log.Printf("cannot get active hours: %s", err.Error())
		doAvailability(hours.Availability.Topic, false)
		return
	}
	doAvailability(hours.Availability.Topic, true)
	doState(hours.StateTopic, fmt.Sprintf("%d", activeHours))
	doSubscribeChanges(baseTopic)
}

func NumberBrightness(monitor Device) {
	objectId := fmt.Sprintf("%s_brightness", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/number/%s", config.CFG.HassDiscoveryPrefix, objectId)
	brightness := Number{
		Name:         "Brightness",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		CommandTopic: fmt.Sprintf("%s/set", baseTopic),
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:brightness-percent",
		Min:          1,
		Max:          100,
		Mode:         "slider",
		Step:         1,
		Unit:         "%",
	}

	doDiscovery(brightness, baseTopic)

	brightnessLevel, err := device.GetBrightnessLevel()
	if err != nil {
		log.Printf("cannot get brightness level: %s", err.Error())
		doAvailability(brightness.Availability.Topic, false)
		return
	}
	doAvailability(brightness.Availability.Topic, true)
	doState(brightness.StateTopic, fmt.Sprintf("%d", brightnessLevel))
	doSubscribeChanges(baseTopic)
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

func doSubscribeChanges(baseTopic string) {
	if token := mqtt.C.Subscribe(fmt.Sprintf("%s/#", baseTopic), 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe: %s", token.Error())
	}
}
