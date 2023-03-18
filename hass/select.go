package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (entity *Select) GetType() string {
	return TypeNumber
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

	discoveryTopic := fmt.Sprintf("%s/config", entity.BaseTopic)

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, false, js)
	if pubToken.Error() != nil {
		log.Fatalf("[%s] failed to publish discovery: %s", entity.ObjectId, pubToken.Error())
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

func (entity *Select) publishState(state string) {
	log.Printf("[%s] publishing state: %s", entity.ObjectId, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, state)
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
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
	listener := func(client mqttLib.Client, msg mqttLib.Message) {
		if msg.Topic() != entity.CommandTopic {
			return
		}
		set := string(msg.Payload())
		err := entity.SetValue(set)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}
	}
	if token := mqtt.C.Subscribe(entity.CommandTopic, 0, listener); token.Wait() && token.Error() != nil {

		return token.Error()
	}

	return nil
}
