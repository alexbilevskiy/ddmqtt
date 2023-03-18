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
}

type NumberEntity interface {
	GetType() string
	SetState(int)
}

type Sensor struct {
	Name         string        `json:"name"`
	Availability SAvailability `json:"availability"`
	StateTopic   string        `json:"state_topic"`
	ObjectId     string        `json:"object_id"`
	UniqueId     string        `json:"unique_id"`
	Device       Device        `json:"device"`
}

func (Sensor) GetType() string {
	return TypeSensor
}

type Number struct {
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

func (Number) GetType() string {
	return TypeNumber
}

type SAvailability struct {
	Topic string `json:"topic"`
}
