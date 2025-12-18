package hass

import (
	"encoding/json"
	"log"
)

type Button struct {
	Discovered     bool   `json:"-"`
	Avaialable     bool   `json:"-"`
	BaseTopic      string `json:"-"`
	DiscoveryTopic string `json:"-"`
	valueSetter    ButtonSetter
	mqtt           mqttClient
	Name           string        `json:"name"`
	Availability   SAvailability `json:"availability"`
	CommandTopic   string        `json:"command_topic"`
	ObjectId       string        `json:"object_id"`
	UniqueId       string        `json:"unique_id"`
	Device         *Device       `json:"device"`
	Icon           string        `json:"icon"`
}

func (entity *Button) SetMqtt(mqtt mqttClient) {
	entity.mqtt = mqtt
}

func (entity *Button) SetValueSetter(setter ButtonSetter) {
	entity.valueSetter = setter
}

func (entity *Button) Init() error {
	entity.DoDiscovery()
	err := entity.subscribeMqtt()
	if err != nil {
		return err
	}

	return nil
}

func (entity *Button) DoDiscovery() {
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

func (entity *Button) ReportValue() error {
	entity.reportAvailability(true)

	return nil
}

func (entity *Button) reportAvailability(available bool) {
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

func (entity *Button) SetValue() error {
	err := entity.valueSetter(entity.Device.Identifiers)
	if err != nil {

		return err
	}

	return nil
}

func (entity *Button) subscribeMqtt() error {
	entity.mqtt.AddListener(entity.CommandTopic, func(payload []byte) {
		set := string(payload)
		if set != "PRESS" {
			log.Printf("[%s] invalid command received: %s", entity.ObjectId, set)
			return
		}
		err := entity.SetValue()
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}
	})

	return nil
}
