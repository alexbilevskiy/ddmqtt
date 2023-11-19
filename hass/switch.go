package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
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

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	pubToken := mqtt.C.Publish(entity.DiscoveryTopic, 0, true, js)
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

func (entity *Switch) publishState(state string) {
	if entity.State == state {
		return
	}
	log.Printf("[%s] publishing state: %s", entity.ObjectId, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, state)
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
	entity.State = state
}

func (entity *Switch) SetValue(value string) error {

	return entity.valueSetter(value)
}

func (entity *Switch) subscribeMqtt() error {
	mqtt.AddListener(entity.CommandTopic, func(client mqttLib.Client, msg mqttLib.Message) {
		if msg.Topic() != entity.CommandTopic {
			return
		}
		set := string(msg.Payload())
		err := entity.SetValue(set)
		if err != nil {
			log.Printf("[%s] failed to set value: %s", entity.ObjectId, err.Error())
		}

		entity.publishState(set)
	})

	return nil
}
