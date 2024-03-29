package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (entity *Select) GetType() string {
	return TypeSelect
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

	pubToken := mqtt.C.Publish(entity.DiscoveryTopic, 0, true, js)
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

func (entity *Select) publishState(state string) {
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
