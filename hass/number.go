package hass

import (
	"ddmqtt/config"
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	mqttLib "github.com/eclipse/paho.mqtt.golang"
	"log"
	"regexp"
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

	pubToken := mqtt.C.Publish(discoveryTopic, 0, false, js)
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

func (entity *Number) publishState(state int) {
	log.Printf("[%s] publishing state: %d", entity.ObjectId, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, strconv.Itoa(state))
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
}

func (entity *Number) SetValue(value int) error {

	return entity.valueSetter(value)
}

func (entity *Number) subscribeMqtt() error {
	listener := func(client mqttLib.Client, msg mqttLib.Message) {
		r, _ := regexp.Compile(fmt.Sprintf("%s/(%s)/([a-zA-Z0-9_-]+)/set", config.CFG.HassDiscoveryPrefix, TypeNumber))
		matches := r.FindStringSubmatch(msg.Topic())
		if matches == nil {
			//log.Printf("skipping mqtt topic: %s with payload `%s`", msg.Topic(), msg.Payload())
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
	}
	if token := mqtt.C.Subscribe(entity.CommandTopic, 0, listener); token.Wait() && token.Error() != nil {

		return token.Error()
	}

	return nil
}