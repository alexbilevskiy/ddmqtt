package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
)

func (entity *Number) GetType() string {
	return TypeNumber
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

	discoveryTopic := fmt.Sprintf("%s/config", entity.BaseTopic)

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	pubToken := mqtt.C.Publish(discoveryTopic, 0, true, js)
	if pubToken.Error() != nil {
		log.Fatalf("[%s] failed to publish discovery: %s", entity.ObjectId, pubToken.Error())
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

	pubOnlineToken := mqtt.C.Publish(entity.Availability.Topic, 0, false, availabilityStatus)
	if pubOnlineToken.Error() != nil {
		log.Fatalf("[%s] failed to publish online state: %s", entity.ObjectId, pubOnlineToken.Error())
	}
	entity.Avaialable = available
}

func (entity *Number) publishState(state int) {
	if entity.State == state {
		return
	}
	log.Printf("[%s] publishing state: %d", entity.ObjectId, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, strconv.Itoa(state))
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
	entity.State = state
}

func (entity *Number) SetValue(value int) error {

	return entity.valueSetter(value)
}

func (entity *Number) subscribeMqtt() error {
	mqtt.AddListener(entity.CommandTopic, func(client mqttLib.Client, msg mqttLib.Message) {
		if msg.Topic() != entity.CommandTopic {
			return
		}
		set := string(msg.Payload())
		value, err := strconv.Atoi(set)
		if err != nil {
			log.Printf("[%s] invalid set value: %s", entity.ObjectId, msg.Payload())
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
