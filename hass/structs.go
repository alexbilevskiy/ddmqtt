package hass

import "ddmqtt/config"

type Device struct {
	Identifiers  string `json:"identifiers"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Name         string `json:"name"`
	SwVersion    string `json:"sw_version"`
}

const TypeSensor = "sensor"
const TypeNumber = "number"
const TypeSelect = "select"
const TypeButton = "button"
const TypeSwitch = "switch"

type NumberEntity interface {
	GetType() string
	Init() error
	DoDiscovery()
	ReportValue() error
	SetValueReader(func() (int, error))
	SetValueSetter(func(value int) error)
	SetValue(value int) error
	reportAvailability(available bool)
	publishState(state int)
	subscribeMqtt() error
}

type Number struct {
	Discovered   bool
	BaseTopic    string
	valueReader  func() (int, error)
	valueSetter  func(value int) error
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	StateTopic   string        `json:"state_topic"`
	CommandTopic string        `json:"command_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
	Icon         string        `json:"icon"`
	Min          int           `json:"min"`
	Max          int           `json:"max"`
	Mode         string        `json:"mode"`
	Step         int           `json:"step"`
	Unit         string        `json:"unit_of_measurement"`
}

type SAvailability struct {
	Topic string `json:"topic"`
}

type SensorEntity interface {
	GetType() string
	Init() error
	DoDiscovery()
	ReportValue() error
	SetValueReader(func() (int, error))
	reportAvailability(available bool)
	publishState(state int)
}

type Sensor struct {
	Discovered   bool
	BaseTopic    string
	valueReader  func() (int, error)
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	StateTopic   string        `json:"state_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
	Icon         string        `json:"icon"`
}

type SelectEntity interface {
	GetType() string
	Init() error
	DoDiscovery()
	ReportValue() error
	SetValueSetter(func(value string) error)
	SetValue(value string) error
	reportAvailability(available bool)
	publishState(state string)
	subscribeMqtt() error
}

type Select struct {
	Discovered   bool
	BaseTopic    string
	valueReader  func() (string, error)
	valueSetter  func(value string) error
	State        string
	Presets      []config.Preset
	Affected     []Number
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	StateTopic   string        `json:"state_topic"`
	CommandTopic string        `json:"command_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
	Options      []string      `json:"options"`
	Icon         string        `json:"icon"`
}

type SwitchEntity interface {
	GetType() string
	Init() error
	DoDiscovery()
	ReportValue() error
	SetValueReader(func() (int, error))
	SetValueSetter(func(value string) error)
	SetValue(value string) error
	reportAvailability(available bool)
	publishState(state string)
	subscribeMqtt() error
}

type Switch struct {
	Discovered   bool
	BaseTopic    string
	valueReader  func() (string, error)
	valueSetter  func(value string) error
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	StateTopic   string        `json:"state_topic"`
	CommandTopic string        `json:"command_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
	Icon         string        `json:"icon"`
}

type ButtonEntity interface {
	GetType() string
	Init() error
	DoDiscovery()
	ReportValue() error
	SetValueSetter(func() error)
	SetValue() error
	reportAvailability(available bool)
	subscribeMqtt() error
}

type Button struct {
	Discovered   bool
	BaseTopic    string
	valueSetter  func() error
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	CommandTopic string        `json:"command_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
	Icon         string        `json:"icon"`
}
