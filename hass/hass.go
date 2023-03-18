package hass

import (
	"ddmqtt/config"
	"ddmqtt/device"
	"ddmqtt/mqtt"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
	"regexp"
	"strconv"
	"time"
)

func StartReporting() {
	var monitor Device

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

	ah := CreateSensorActiveHours(monitor)
	ah.DoDiscovery()

	br := CreateNumberBrightness(monitor)
	br.DoDiscovery()
	subscribeChanges(br)

	cn := CreateNumberContrast(monitor)
	cn.DoDiscovery()
	subscribeChanges(cn)

	for {
		var err error
		err = ah.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", ah.ObjectId)
		}
		time.Sleep(200 * time.Millisecond)
		err = br.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", br.ObjectId)
		}
		time.Sleep(200 * time.Millisecond)
		err = cn.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", cn.ObjectId)
		}

		time.Sleep(60 * time.Second)
	}
}

func CreateSensorActiveHours(monitor Device) Sensor {
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

	hours.SetValueReader(device.GetMonitorActiveHours)

	return hours
}

func CreateNumberBrightness(monitor Device) Number {
	objectId := fmt.Sprintf("%s_brightness", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/number/%s", config.CFG.HassDiscoveryPrefix, objectId)
	brightness := Number{
		Discovered:   false,
		BaseTopic:    baseTopic,
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
	brightness.SetValueReader(device.GetBrightnessLevel)
	brightness.SetValueSetter(device.SetBrightnessLevel)

	return brightness
}

func CreateNumberContrast(monitor Device) Number {
	objectId := fmt.Sprintf("%s_contrast", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/number/%s", config.CFG.HassDiscoveryPrefix, objectId)
	contrast := Number{
		Discovered:   false,
		BaseTopic:    baseTopic,
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

	contrast.SetValueReader(device.GetContrastLevel)
	contrast.SetValueSetter(device.SetContrastLevel)

	return contrast
}

func subscribeChanges(entity Number) {
	listener := func(client mqttLib.Client, msg mqttLib.Message) {
		r, _ := regexp.Compile(fmt.Sprintf("%s/(%s)/([a-zA-Z0-9_-]+)/set", config.CFG.HassDiscoveryPrefix, TypeNumber))
		matches := r.FindStringSubmatch(msg.Topic())
		if matches == nil {
			//log.Printf("skipping mqtt topic: %s with payload `%s`", msg.Topic(), msg.Payload())
			return
		}
		set := string(msg.Payload())
		value, err := strconv.Atoi(set)
		if err != nil {
			log.Printf("[%s] invalid set value: %s", entity.ObjectId, msg.Payload())
			return
		}
		err = entity.SetValue(value)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}

	}
	if token := mqtt.C.Subscribe(fmt.Sprintf("%s/#", entity.BaseTopic), 0, listener); token.Wait() && token.Error() != nil {
		log.Fatalf("[%s] failed to subscribe: %s", entity.ObjectId, token.Error())
	}
}
