package hass

import (
	"ddmqtt/config"
	"ddmqtt/ddmrpc"
	"fmt"
	"log"
	"time"
)

func StartReporting() {
	var monitor Device

	attrs, err := ddmrpc.GetAssetAttributes()
	if err != nil {
		log.Fatalf("failed to read monitor: %s", err.Error())
	}
	fw, err := ddmrpc.GetFirmwareVersion()
	if err != nil {
		log.Fatalf("failed to read monitor fw: %s", err.Error())
	}
	monitor = Device{
		Identifiers:  attrs.ServiceTag,
		Manufacturer: "Dell",
		Model:        attrs.ModelCode,
		Name:         attrs.Model,
		SwVersion:    fw,
	}

	ah := CreateSensorActiveHours(monitor)
	ah.Init()

	br := CreateNumberBrightness(monitor)
	err = br.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	cn := CreateNumberContrast(monitor)
	err = cn.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	se := CreateSelectPresets(monitor)
	err = se.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}
	se.Affected = append(se.Affected, br, cn)

	pw := CreateSwitchPower(monitor)
	err = pw.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	re := CreateButtonReset(monitor)
	err = re.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	for {
		var err error
		err = ah.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", ah.ObjectId)
		}
		err = br.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", br.ObjectId)
		}
		err = cn.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", cn.ObjectId)
		}

		err = se.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", cn.ObjectId)
		}

		err = pw.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", cn.ObjectId)
		}

		err = re.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to report state", cn.ObjectId)
		}

		time.Sleep(15 * time.Second)
	}
}

func CreateSensorActiveHours(monitor Device) Sensor {
	objectId := fmt.Sprintf("%s_active_hours", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/sensor/%s", config.CFG.HassDiscoveryPrefix, objectId)
	hours := Sensor{
		Discovered:   false,
		BaseTopic:    baseTopic,
		Name:         "Active hours",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:clock-outline",
	}

	hours.SetValueReader(ddmrpc.GetMonitorActiveHours)

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
		Min:          0,
		Max:          100,
		Mode:         "slider",
		Step:         1,
		Unit:         "%",
	}
	brightness.SetValueReader(ddmrpc.GetBrightnessLevel)
	brightness.SetValueSetter(ddmrpc.SetBrightnessLevel)

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
		Min:          0,
		Max:          100,
		Mode:         "slider",
		Step:         1,
		Unit:         "%",
	}

	contrast.SetValueReader(ddmrpc.GetContrastLevel)
	contrast.SetValueSetter(ddmrpc.SetContrastLevel)

	return contrast
}

func CreateSelectPresets(monitor Device) Select {
	objectId := fmt.Sprintf("%s_presets", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/select/%s", config.CFG.HassDiscoveryPrefix, objectId)
	selector := Select{
		Discovered:   false,
		BaseTopic:    baseTopic,
		Name:         "Preset",
		State:        "",
		Presets:      config.CFG.Presets,
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		CommandTopic: fmt.Sprintf("%s/set", baseTopic),
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:format-list-bulleted",
		Options:      append(make([]string, 0), ""),
		Affected:     make([]Number, 0),
	}
	for _, option := range selector.Presets {
		selector.Options = append(selector.Options, option.Name)
	}

	selector.SetValueReader(func() (string, error) {
		return selector.State, nil
	})
	selector.SetValueSetter(func(value string) error {
		found := false
		var err error
		for _, option := range selector.Presets {
			if option.Name != value {

				continue
			}
			found = true
			err = ddmrpc.SetBrightnessLevel(option.Brightness)
			if err != nil {
				return err
			}
			err = ddmrpc.SetContrastLevel(option.Contrast)
			if err != nil {
				return err
			}
			selector.State = value
		}
		if !found {
			log.Printf("[%s] not found option `%s`", selector.ObjectId, value)
		}

		return nil
	})

	return selector
}

func CreateSwitchPower(monitor Device) Switch {
	objectId := fmt.Sprintf("%s_power", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/switch/%s", config.CFG.HassDiscoveryPrefix, objectId)
	power := Switch{
		Discovered:   false,
		BaseTopic:    baseTopic,
		Name:         "Power",
		StateTopic:   fmt.Sprintf("%s/state", baseTopic),
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		CommandTopic: fmt.Sprintf("%s/set", baseTopic),
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:power",
	}

	power.SetValueReader(ddmrpc.GetPower)
	power.SetValueSetter(ddmrpc.SetPower)

	return power
}

func CreateButtonReset(monitor Device) Button {
	objectId := fmt.Sprintf("%s_reset", monitor.Identifiers)
	baseTopic := fmt.Sprintf("%s/button/%s", config.CFG.HassDiscoveryPrefix, objectId)
	power := Button{
		Discovered:   false,
		BaseTopic:    baseTopic,
		Name:         "Reset",
		Availability: SAvailability{Topic: fmt.Sprintf("%s/available", baseTopic)},
		CommandTopic: fmt.Sprintf("%s/press", baseTopic),
		ObjectId:     objectId,
		UniqueId:     objectId,
		Device:       monitor,
		Icon:         "mdi:restart",
	}

	power.SetValueSetter(ddmrpc.Reset)

	return power
}
