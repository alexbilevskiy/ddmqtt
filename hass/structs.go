package hass

type Device struct {
	Identifiers  string `json:"identifiers"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Name         string `json:"name"`
}

const TypeSensor = "sensor"
const TypeNumber = "number"

type Entity interface {
	GetType() string
	DoDiscovery()
	DoAvailability(available bool)
	ReportState(state int)
	ReportValue() error
	SetValueReader(func() (int, error))
}

type NumberEntity interface {
	GetType() string
	SetValueSetter(func(value int) error)
	SetValue(value int) error
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
