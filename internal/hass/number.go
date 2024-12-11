package hass

import (
	"encoding/json"
	"log"
	"strconv"
)

type Number struct {
	Discovered     bool   `json:"-"`
	Avaialable     bool   `json:"-"`
	State          int    `json:"-"`
	BaseTopic      string `json:"-"`
	DiscoveryTopic string `json:"-"`
	valueReader    func() (int, error)
	valueSetter    func(value int) error
	mqtt           mqttClient
	Name           string        `json:"name"`
	Availability   SAvailability `json:"availability"`
	StateTopic     string        `json:"state_topic"`
	CommandTopic   string        `json:"command_topic"`
	ObjectId       string        `json:"object_id"`
	UniqueId       string        `json:"unique_id"`
	Device         *Device       `json:"device"`
	Icon           string        `json:"icon"`
	Min            int           `json:"min"`
	Max            int           `json:"max"`
	Mode           string        `json:"mode"`
	Step           int           `json:"step"`
	Unit           string        `json:"unit_of_measurement"`
}

func (entity *Number) SetMqtt(mqtt mqttClient) {
	entity.mqtt = mqtt
}

func (entity *Number) SetValueReader(reader func() (int, error)) {
	entity.valueReader = reader
}

func (entity *Number) SetValueSetter(setter func(value int) error) {
	entity.valueSetter = setter
}

func (entity *Number) Init() error {
	entity.DoDiscovery()
	err := entity.subscribeMqtt()
	if err != nil {
		return err
	}

	return nil
}

func (entity *Number) DoDiscovery() {
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

func (entity *Number) ReportValue() error {
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

func (entity *Number) reportAvailability(available bool) {
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
		log.Printf("[%s] failed to publish online state: %s", entity.ObjectId, err.Error())
	}
	entity.Avaialable = available
}

func (entity *Number) publishState(state int) {
	if entity.State == state {
		return
	}
	log.Printf("[%s] publishing state: %d", entity.ObjectId, state)

	err := entity.mqtt.Publish(entity.StateTopic, false, strconv.Itoa(state))
	if err != nil {
		log.Printf("[%s] failed to publish state data: %s", entity.ObjectId, err.Error())
	}
	entity.State = state
}

func (entity *Number) SetValue(value int) error {

	return entity.valueSetter(value)
}

func (entity *Number) subscribeMqtt() error {
	entity.mqtt.AddListener(entity.CommandTopic, func(payload []byte) {
		set := string(payload)
		value, err := strconv.Atoi(set)
		if err != nil {
			log.Printf("[%s] invalid set value: %s", entity.ObjectId, payload)
			return
		}
		err = entity.SetValue(value)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}

		entity.publishState(value)
	})

	return nil
}
