package hass

import (
	"encoding/json"
	"log"
	"strconv"
)

type Number struct {
	Discovered       bool   `json:"-"`
	Avaialable       bool   `json:"-"`
	State            int    `json:"-"`
	BaseTopic        string `json:"-"`
	DiscoveryTopic   string `json:"-"`
	valueReader      NumberReader
	valueSetter      NumberSetter
	mqtt             mqttClient
	Name             string          `json:"name"`
	Availability     []SAvailability `json:"availability"`
	AvailabilityMode string          `json:"availability_mode"`
	StateTopic       string          `json:"state_topic"`
	CommandTopic     string          `json:"command_topic"`
	ObjectId         string          `json:"object_id"`
	UniqueId         string          `json:"unique_id"`
	Device           *Device         `json:"device"`
	Icon             string          `json:"icon"`
	Min              int             `json:"min"`
	Max              int             `json:"max"`
	Mode             string          `json:"mode"`
	Step             int             `json:"step"`
	Unit             string          `json:"unit_of_measurement"`
}

func (entity *Number) SetMqtt(mqtt mqttClient) {
	entity.mqtt = mqtt
}

func (entity *Number) SetValueReader(reader NumberReader) {
	entity.valueReader = reader
}

func (entity *Number) SetValueSetter(setter NumberSetter) {
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
	value, err := entity.valueReader(entity.Device.Identifiers)
	if err != nil {
		log.Printf("[%s] cannot read value: %s", entity.ObjectId, err.Error())

		return err
	}
	entity.publishState(value)

	return nil
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

	return entity.valueSetter(entity.Device.Identifiers, value)
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
