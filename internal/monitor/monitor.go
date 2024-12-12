package monitor

import (
	"fmt"
	"log"
	"time"

	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/hass"
	"ddmqtt/internal/mqtt"
)

type DDMRPCClient interface {
	GetAssetAttributes() (ddmrpc.AssetAttributes, error)
	GetCapabilities() (ddmrpc.Capabilities, error)
	GetFirmwareVersion() (string, error)
	GetMonitorActiveHours() (int, error)
	GetBrightnessLevel() (int, error)
	SetBrightnessLevel(brightness int) error
	GetContrastLevel() (int, error)
	SetContrastLevel(contrast int) error
	GetPower() (string, error)
	SetPower(value string) error
	GetActiveInput() (byte, error)
	SetActiveInput(input byte) error
	Reset() error
}
type Monitor struct {
	cfg          *config.Config
	ddmrpc       DDMRPCClient
	mqtt         *mqtt.Client
	device       *hass.Device
	capabilities ddmrpc.Capabilities
}

func NewMonitor(cfg *config.Config, ddmrpc DDMRPCClient, mqtt *mqtt.Client) Monitor {
	var device hass.Device

	attrs, err := ddmrpc.GetAssetAttributes()
	if err != nil {
		log.Fatalf("failed to read monitor: %s", err.Error())
	}
	fw, err := ddmrpc.GetFirmwareVersion()
	if err != nil {
		log.Fatalf("failed to read monitor fw: %s", err.Error())
	}
	caps, err := ddmrpc.GetCapabilities()
	if err != nil {
		log.Fatalf("failed to read monitor caps: %s", err.Error())
	}
	log.Printf("found monitor: %v", attrs)
	device = hass.Device{
		Identifiers:  attrs.ServiceTag,
		Manufacturer: "Dell",
		Model:        attrs.ModelCode,
		Name:         attrs.Model,
		SwVersion:    fw,
	}

	return Monitor{
		cfg:          cfg,
		ddmrpc:       ddmrpc,
		mqtt:         mqtt,
		device:       &device,
		capabilities: caps,
	}
}

func (m *Monitor) CreateSensorActiveHours() hass.Sensor {
	objectId := fmt.Sprintf("%s_active_hours", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	hours := hass.Sensor{
		BaseTopic:      baseTopic,
		Name:           "Active hours",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/sensor/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:clock-outline",
	}

	hours.SetMqtt(m.mqtt)
	hours.SetValueReader(m.ddmrpc.GetMonitorActiveHours)

	return hours
}

func (m *Monitor) CreateNumberBrightness() hass.Number {
	objectId := fmt.Sprintf("%s_brightness", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	brightness := hass.Number{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Brightness",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/number/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/set", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:brightness-percent",
		Min:            0,
		Max:            100,
		Mode:           "slider",
		Step:           1,
		Unit:           "%",
	}

	brightness.SetMqtt(m.mqtt)
	brightness.SetValueReader(m.ddmrpc.GetBrightnessLevel)
	brightness.SetValueSetter(m.ddmrpc.SetBrightnessLevel)

	return brightness
}

func (m *Monitor) CreateNumberContrast() hass.Number {
	objectId := fmt.Sprintf("%s_contrast", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	contrast := hass.Number{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Contrast",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/number/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/set", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:contrast-box",
		Min:            0,
		Max:            100,
		Mode:           "slider",
		Step:           1,
		Unit:           "%",
	}

	contrast.SetMqtt(m.mqtt)
	contrast.SetValueReader(m.ddmrpc.GetContrastLevel)
	contrast.SetValueSetter(m.ddmrpc.SetContrastLevel)

	return contrast
}

func (m *Monitor) CreateSelectPresets() hass.Select {
	objectId := fmt.Sprintf("%s_presets", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	selector := hass.Select{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Preset",
		State:          "",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/select/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/set", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:format-list-bulleted",
		Options:        append(make([]string, 0), ""),
		Affected:       make([]hass.Number, 0),
	}
	for _, option := range m.cfg.Presets {
		selector.Options = append(selector.Options, option.Name)
	}

	selector.SetMqtt(m.mqtt)
	selector.SetValueReader(func() (string, error) {
		return selector.State, nil
	})
	selector.SetValueSetter(func(value string) error {
		found := false
		var err error
		for _, option := range m.cfg.Presets {
			if option.Name != value {

				continue
			}
			found = true
			err = m.ddmrpc.SetBrightnessLevel(option.Brightness)
			if err != nil {
				return err
			}
			err = m.ddmrpc.SetContrastLevel(option.Contrast)
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

func (m *Monitor) CreateSelectInput() hass.Select {
	objectId := fmt.Sprintf("%s_input", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	selector := hass.Select{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Input",
		State:          "",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/select/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/set", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:import",
		Options:        append(make([]string, 0), ""),
		Affected:       make([]hass.Number, 0),
	}
	for _, option := range m.capabilities.AvailableInputs {
		if _, ok := ddmrpc.KnownInputs[option]; !ok {
			continue
		}
		selector.Options = append(selector.Options, ddmrpc.KnownInputs[option])
	}

	selector.SetMqtt(m.mqtt)
	selector.SetValueReader(func() (string, error) {
		input, err := m.ddmrpc.GetActiveInput()
		if err != nil {
			return "", err
		}
		_, ok := ddmrpc.KnownInputs[input]
		if !ok {
			log.Printf("[%s] unknown input `%x`", selector.ObjectId, input)
			return "", err
		}
		return ddmrpc.KnownInputs[input], nil
	})
	selector.SetValueSetter(func(value string) error {
		found := false
		var err error
		for input, option := range ddmrpc.KnownInputs {
			if option != value {

				continue
			}
			found = true
			err = m.ddmrpc.SetActiveInput(input)
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

func (m *Monitor) CreateSwitchPower() hass.Switch {
	objectId := fmt.Sprintf("%s_power", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	power := hass.Switch{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Power",
		StateTopic:     fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic: fmt.Sprintf("%s/switch/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/set", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:power",
	}

	power.SetMqtt(m.mqtt)
	power.SetValueReader(m.ddmrpc.GetPower)
	power.SetValueSetter(m.ddmrpc.SetPower)

	return power
}

func (m *Monitor) CreateButtonReset() hass.Button {
	objectId := fmt.Sprintf("%s_reset", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	reset := hass.Button{
		Discovered:     false,
		BaseTopic:      baseTopic,
		Name:           "Reset",
		DiscoveryTopic: fmt.Sprintf("%s/button/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:   hass.SAvailability{Topic: fmt.Sprintf("%s/%s/available", m.cfg.MqttRootTopic, m.device.Identifiers)},
		CommandTopic:   fmt.Sprintf("%s/press", baseTopic),
		ObjectId:       objectId,
		UniqueId:       objectId,
		Device:         m.device,
		Icon:           "mdi:restart",
	}

	reset.SetMqtt(m.mqtt)
	reset.SetValueSetter(m.ddmrpc.Reset)

	return reset
}

func (m *Monitor) StartReporting() {
	var err error

	m.mqtt.Connect(m.device.Identifiers)

	ah := m.CreateSensorActiveHours()
	ah.Init()

	br := m.CreateNumberBrightness()
	err = br.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	cn := m.CreateNumberContrast()
	err = cn.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	se := m.CreateSelectPresets()
	err = se.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}
	se.Affected = append(se.Affected, br, cn)

	si := m.CreateSelectInput()
	err = si.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	pw := m.CreateSwitchPower()
	err = pw.Init()
	if err != nil {
		log.Fatalf("[%s] failed to init: %s", br.ObjectId, err.Error())
	}

	re := m.CreateButtonReset()
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

		err = si.ReportValue()
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
