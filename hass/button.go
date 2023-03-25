package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (entity *Button) GetType() string {
	return TypeButton
}

func (entity *Button) SetValueSetter(setter func() error) {
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

	discoveryTopic := fmt.Sprintf("%s/config", entity.BaseTopic)

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, true, js)
	if pubToken.Error() != nil {
		log.Fatalf("[%s] failed to publish discovery: %s", entity.ObjectId, pubToken.Error())
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

	pubOnlineToken := mqtt.C.Publish(entity.Availability.Topic, 0, false, availabilityStatus)
	if pubOnlineToken.Error() != nil {
		log.Fatalf("[%s] failed to publish online state: %s", entity.ObjectId, pubOnlineToken.Error())
	}
	entity.Avaialable = available
}

func (entity *Button) SetValue() error {
	err := entity.valueSetter()
	if err != nil {

		return err
	}

	return nil
}

func (entity *Button) subscribeMqtt() error {
	listener := func(client mqttLib.Client, msg mqttLib.Message) {
		if msg.Topic() != entity.CommandTopic {
			return
		}
		set := string(msg.Payload())
		if set != "PRESS" {
			log.Printf("[%s] invalid command received: %s", entity.ObjectId, set)
			return
		}
		err := entity.SetValue()
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}
	}
	if token := mqtt.C.Subscribe(entity.CommandTopic, 0, listener); token.Wait() && token.Error() != nil {

		return token.Error()
	}

	return nil
}
