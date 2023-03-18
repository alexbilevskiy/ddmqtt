package hass

import (
	"ddmqtt/config"
	"ddmqtt/device"
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var discovered = make(map[string]Entity, 0)
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
	NumberContrast(monitor)
}

func doDiscovery(entity Entity, baseTopic string) {
	var objectId string
	switch entity.GetType() {
	case TypeSensor:
		objectId = entity.(Sensor).ObjectId
	case TypeNumber:
		objectId = entity.(Number).ObjectId
	}
	if _, ok := discovered[objectId]; ok {
		return
	}
	discovered[objectId] = entity

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
	doSubscribeChanges(brightness.CommandTopic)
}

func NumberContrast(monitor Device) {
	objectId := fmt.Sprintf("%s_contrast", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/number/%s", config.CFG.HassDiscoveryPrefix, objectId)
	contrast := Number{
		Name:         "Contrast",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		CommandTopic: fmt.Sprintf("%s/set", baseTopic),
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:contrast-box",
		Min:          1,
		Max:          100,
		Mode:         "slider",
		Step:         1,
		Unit:         "%",
	}

	doDiscovery(contrast, baseTopic)

	contrastLevel, err := device.GetContrastLevel()
	if err != nil {
		log.Printf("cannot get contrast level: %s", err.Error())
		doAvailability(contrast.Availability.Topic, false)
		return
	}
	doAvailability(contrast.Availability.Topic, true)
	doState(contrast.StateTopic, fmt.Sprintf("%d", contrastLevel))
	doSubscribeChanges(contrast.CommandTopic)
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
	if token := mqtt.C.Subscribe(fmt.Sprintf("%s/#", baseTopic), 0, mqttListener); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to subscribe: %s", token.Error())
	}
}

var mqttListener mqttLib.MessageHandler = func(client mqttLib.Client, msg mqttLib.Message) {
	r, _ := regexp.Compile(fmt.Sprintf("%s/(%s)/([a-zA-Z0-9_-]+)/set", config.CFG.HassDiscoveryPrefix, TypeNumber))
	matches := r.FindStringSubmatch(msg.Topic())
	if matches == nil {
		log.Printf("skipping mqtt topic: %s with payload `%s`", msg.Topic(), msg.Payload())
		return
	}
	objectId := matches[2]
	if _, ok := discovered[objectId]; !ok {
		log.Printf("no such device to set value: %s", objectId)
		return
	}
	set := string(msg.Payload())
	value, err := strconv.Atoi(set)
	if err != nil {
		log.Printf("invalid set value: %s", msg.Payload())
		return
	}
	if strings.Contains(objectId, "brightness") {
		err = device.SetBrightnessLevel(value)
	} else if strings.Contains(objectId, "contrast") {
		err = device.SetContrastLevel(value)
	} else {
		log.Fatalf("invalid device to set: %s", objectId)
	}
	if err != nil {
		log.Printf("failed to set value: %s", err.Error())
	}
	entity := discovered[objectId].(Number)
	doState(entity.StateTopic, set)
}
