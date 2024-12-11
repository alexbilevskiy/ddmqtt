package hass

import (
	"ddmqtt/internal/config"
	"encoding/json"
	"log"
)

type Select struct {
	Discovered     bool   `json:"-"`
	Avaialable     bool   `json:"-"`
	State          string `json:"-"`
	BaseTopic      string `json:"-"`
	DiscoveryTopic string `json:"-"`
	valueReader    func() (string, error)
	valueSetter    func(value string) error
	mqtt           mqttClient
	Presets        []config.Preset `json:"-"`
	Affected       []Number        `json:"-"`
	Name           string          `json:"name"`
	Availability   SAvailability   `json:"availability"`
	StateTopic     string          `json:"state_topic"`
	CommandTopic   string          `json:"command_topic"`
	ObjectId       string          `json:"object_id"`
	UniqueId       string          `json:"unique_id"`
	Device         *Device         `json:"device"`
	Options        []string        `json:"options"`
	Icon           string          `json:"icon"`
}

func (entity *Select) SetMqtt(mqtt mqttClient) {
	entity.mqtt = mqtt
}

func (entity *Select) SetValueReader(reader func() (string, error)) {
	entity.valueReader = reader
}

func (entity *Select) SetValueSetter(setter func(value string) error) {
	entity.valueSetter = setter
}

func (entity *Select) Init() error {
	entity.DoDiscovery()
	err := entity.subscribeMqtt()
	if err != nil {
		return err
	}

	return nil
}

func (entity *Select) DoDiscovery() {
	if entity.Discovered {
		return
	}

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	err := entity.mqtt.Publish(entity.DiscoveryTopic, true, js)
	if err != nil {
		log.Printf("[%s] failed to publish discovery: %s", entity.ObjectId, err.Error())
	}
}

func (entity *Select) ReportValue() error {
	value, err := entity.valueReader()
	if err != nil {
		log.Printf("[%s] cannot read value: %s", entity.ObjectId, err.Error())
		entity.reportAvailability(false)

		return err
	}
	entity.reportAvailability(true)
	entity.publishState(value)

	return nil
}

func (entity *Select) reportAvailability(available bool) {
	if entity.Avaialable == available {
		return
	}
	availabilityStatus := "offline"
	if available {
		availabilityStatus = "online"
	}
	log.Printf("[%s] publishing availability: %s", entity.ObjectId, availabilityStatus)

	err := entity.mqtt.Publish(entity.Availability.Topic, false, availabilityStatus)
	if err != nil {
		log.Printf("[%s] failed to publish online state: %s", entity.ObjectId, err)
	}
	entity.Avaialable = available
}

func (entity *Select) publishState(state string) {
	if entity.State == state {
		return
	}
	log.Printf("[%s] publishing state: %s", entity.ObjectId, state)

	err := entity.mqtt.Publish(entity.StateTopic, false, state)
	if err != nil {
		log.Printf("[%s] failed to publish state data: %s", entity.ObjectId, err.Error())
	}
	entity.State = state
}

func (entity *Select) SetValue(value string) error {
	err := entity.valueSetter(value)
	if err != nil {

		return err
	}
	for _, number := range entity.Affected {
		err = number.ReportValue()
		if err != nil {
			log.Printf("[%s] failed to update affected entity: %s", entity.ObjectId, number.ObjectId)
		}
	}

	return nil
}

func (entity *Select) subscribeMqtt() error {
	entity.mqtt.AddListener(entity.CommandTopic, func(payload []byte) {
		set := string(payload)
		err := entity.SetValue(set)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}

		entity.publishState(set)
	})

	return nil
}
