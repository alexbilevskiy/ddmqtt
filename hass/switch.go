package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (entity *Switch) GetType() string {
	return TypeSwitch
}

func (entity *Switch) SetValueReader(reader func() (string, error)) {
	entity.valueReader = reader
}

func (entity *Switch) SetValueSetter(setter func(value string) error) {
	entity.valueSetter = setter
}

func (entity *Switch) Init() error {
	entity.DoDiscovery()
	err := entity.subscribeMqtt()
	if err != nil {
		return err
	}

	return nil
}

func (entity *Switch) DoDiscovery() {
	if entity.Discovered {
		return
	}

	discoveryTopic := fmt.Sprintf("%s/config", entity.BaseTopic)

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, false, js)
	if pubToken.Error() != nil {
		log.Fatalf("[%s] failed to publish discovery: %s", entity.ObjectId, pubToken.Error())
	}
}

func (entity *Switch) ReportValue() error {
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

func (entity *Switch) reportAvailability(available bool) {
	availabilityStatus := "offline"
	if available {
		availabilityStatus = "online"
	}
	log.Printf("[%s] publishing availability: %s", entity.ObjectId, availabilityStatus)

	pubOnlineToken := mqtt.C.Publish(entity.Availability.Topic, 0, false, availabilityStatus)
	if pubOnlineToken.Error() != nil {
		log.Fatalf("[%s] failed to publish online state: %s", entity.ObjectId, pubOnlineToken.Error())
	}
}

func (entity *Switch) publishState(state string) {
	log.Printf("[%s] publishing state: %s", entity.ObjectId, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, state)
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
}

func (entity *Switch) SetValue(value string) error {

	return entity.valueSetter(value)
}

func (entity *Switch) subscribeMqtt() error {
	listener := func(client mqttLib.Client, msg mqttLib.Message) {
		if msg.Topic() != entity.CommandTopic {
			return
		}
		set := string(msg.Payload())
		err := entity.SetValue(set)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}

		entity.publishState(set)
	}
	if token := mqtt.C.Subscribe(entity.CommandTopic, 0, listener); token.Wait() && token.Error() != nil {

		return token.Error()
	}

	return nil
}