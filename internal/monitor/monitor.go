package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/hass"
	"ddmqtt/internal/mqtt"
)

type Monitor struct {
	cfg               *config.Config
	ddmrpc            DDMRPCClient
	mqtt              *mqtt.Client
	device            *hass.Device
	capabilities      ddmrpc.Capabilities
	available         bool
	availabilityTopic string
}

func newMonitor(cfg *config.Config, ddmrpc DDMRPCClient, mqtt *mqtt.Client, attrs *ddmrpc.AssetAttributes) (*Monitor, error) {
	var device hass.Device

	fw, err := ddmrpc.GetFirmwareVersion(attrs.ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("failed to read monitor fw: %w", err)
	}
	caps, err := ddmrpc.GetCapabilities(attrs.ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("failed to read monitor caps: %w", err)
	}
	device = hass.Device{
		Identifiers:  attrs.ServiceTag,
		Manufacturer: "Dell",
		Model:        attrs.ModelCode,
		Name:         attrs.Model,
		SwVersion:    fw,
	}

	return &Monitor{
		cfg:               cfg,
		ddmrpc:            ddmrpc,
		mqtt:              mqtt,
		device:            &device,
		capabilities:      caps,
		availabilityTopic: fmt.Sprintf("%s/%s_available", cfg.MqttRootTopic, device.Identifiers),
	}, nil
}

func (m *Monitor) CreateSensorActiveHours() *hass.Sensor {
	objectId := fmt.Sprintf("%s_active_hours", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	hours := hass.Sensor{
		BaseTopic:        baseTopic,
		Name:             "Active hours",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/sensor/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:clock-outline",
	}

	hours.SetMqtt(m.mqtt)
	hours.SetValueReader(m.ddmrpc.GetMonitorActiveHours)

	return &hours
}

func (m *Monitor) CreateNumberBrightness() *hass.Number {
	objectId := fmt.Sprintf("%s_brightness", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	brightness := hass.Number{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Brightness",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/number/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/set", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:brightness-percent",
		Min:              0,
		Max:              100,
		Mode:             "slider",
		Step:             1,
		Unit:             "%",
	}

	brightness.SetMqtt(m.mqtt)
	brightness.SetValueReader(m.ddmrpc.GetBrightnessLevel)
	brightness.SetValueSetter(m.ddmrpc.SetBrightnessLevel)

	return &brightness
}

func (m *Monitor) CreateNumberContrast() *hass.Number {
	objectId := fmt.Sprintf("%s_contrast", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	contrast := hass.Number{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Contrast",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/number/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/set", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:contrast-box",
		Min:              0,
		Max:              100,
		Mode:             "slider",
		Step:             1,
		Unit:             "%",
	}

	contrast.SetMqtt(m.mqtt)
	contrast.SetValueReader(m.ddmrpc.GetContrastLevel)
	contrast.SetValueSetter(m.ddmrpc.SetContrastLevel)

	return &contrast
}

func (m *Monitor) CreateSelectPresets() *hass.Select {
	objectId := fmt.Sprintf("%s_presets", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	selector := hass.Select{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Preset",
		State:            "",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/select/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/set", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:format-list-bulleted",
		Options:          append(make([]string, 0), ""),
		Affected:         make([]*hass.Number, 0),
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
			err = m.ddmrpc.SetBrightnessLevel(m.device.Identifiers, option.Brightness)
			if err != nil {
				return err
			}
			err = m.ddmrpc.SetContrastLevel(m.device.Identifiers, option.Contrast)
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

	return &selector
}

func (m *Monitor) CreateSelectInput() *hass.Select {
	objectId := fmt.Sprintf("%s_input", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	selector := hass.Select{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Input",
		State:            "",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/select/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/set", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:import",
		Options:          append(make([]string, 0), ""),
		Affected:         make([]*hass.Number, 0),
	}
	for _, option := range m.capabilities.AvailableInputs {
		if _, ok := ddmrpc.KnownInputs[option]; !ok {
			continue
		}
		selector.Options = append(selector.Options, ddmrpc.KnownInputs[option])
	}

	selector.SetMqtt(m.mqtt)
	selector.SetValueReader(func() (string, error) {
		input, err := m.ddmrpc.GetActiveInput(m.device.Identifiers)
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
			err = m.ddmrpc.SetActiveInput(m.device.Identifiers, input)
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

	return &selector
}

func (m *Monitor) CreateSwitchPower() *hass.Switch {
	objectId := fmt.Sprintf("%s_power", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	power := hass.Switch{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Power",
		StateTopic:       fmt.Sprintf("%s/state", baseTopic),
		DiscoveryTopic:   fmt.Sprintf("%s/switch/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/set", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:power",
	}

	power.SetMqtt(m.mqtt)
	power.SetValueReader(func(identifier string) (string, error) {
		pw, err := m.ddmrpc.GetPower(identifier)
		if err != nil {
			return "", fmt.Errorf("get power for %s: %w", identifier, err)
		}
		if pw == "POWER_OFF" {
			return "", fmt.Errorf("device %s is physically off", identifier)
		}
		return pw, nil
	})
	power.SetValueSetter(m.ddmrpc.SetPower)

	return &power
}

func (m *Monitor) CreateButtonReset() *hass.Button {
	objectId := fmt.Sprintf("%s_reset", m.device.Identifiers)
	baseTopic := fmt.Sprintf("%s/%s", m.cfg.MqttRootTopic, objectId)
	reset := hass.Button{
		Discovered:       false,
		BaseTopic:        baseTopic,
		Name:             "Reset",
		DiscoveryTopic:   fmt.Sprintf("%s/button/%s/config", m.cfg.HassDiscoveryPrefix, objectId),
		Availability:     []hass.SAvailability{{Topic: m.availabilityTopic}, {Topic: m.mqtt.GetGlobalAvailabilityTopic()}},
		AvailabilityMode: "all",
		CommandTopic:     fmt.Sprintf("%s/press", baseTopic),
		ObjectId:         objectId,
		UniqueId:         objectId,
		Device:           m.device,
		Icon:             "mdi:restart",
	}

	reset.SetMqtt(m.mqtt)
	reset.SetValueSetter(m.ddmrpc.Reset)

	return &reset
}

type Reporter interface {
	Init() error
	ReportValue() error
}
func (m *Monitor) StartReporting(ctx context.Context) {
	var err error
	var reporters []Reporter

	pw := m.CreateSwitchPower()
	br := m.CreateNumberBrightness()
	cn := m.CreateNumberContrast()
	pr := m.CreateSelectPresets()
	pr.Affected = append(pr.Affected, br, cn)
	reporters = append(reporters, m.CreateSensorActiveHours(), br, cn, pr, m.CreateSelectInput(), m.CreateButtonReset())
	for k, _ := range reporters {
		err = reporters[k].Init()
		if err != nil {
			log.Printf("init reporter err: %v", err)
			//return fmt.Errorf("start monitor %s: %w", m.device.Identifiers, err)
		}
	}

	m.reportAvailability(true)
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()

	firstRun := make(chan struct{}, 1)
	firstRun <- struct{}{}
	for {
		select {
		case <-ctx.Done():
			m.reportAvailability(false)
			return
		case <-t.C:
		case <-firstRun:
			firstRun = nil
		}

		available := true
		for k, _ := range reporters {
			err = pw.ReportValue()
			if err != nil {
				available = false
				break
			}
			err = reporters[k].ReportValue()
			if err != nil {
				available = false
				break
			}
		}
		if available {
			m.reportAvailability(true)

		} else {
			m.reportAvailability(false)
		}
	}
}

func (m *Monitor) reportAvailability(available bool) {
	if m.available == available {
		return
	}
	availabilityStatus := "offline"
	if available {
		availabilityStatus = "online"
	}
	log.Printf("[%s] publishing availability: %s", m.device.Identifiers, availabilityStatus)

	err := m.mqtt.Publish(m.availabilityTopic, false, availabilityStatus)
	if err != nil {
		log.Printf("[%s] failed to publish online state: %s", m.device.Identifiers, err.Error())
	}
	m.available = available
}
