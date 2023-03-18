package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
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
		entity.DoAvailability(false)

		return err
	}
	entity.DoAvailability(true)
	entity.ReportState(value)

	return nil
}

func (entity *Number) DoAvailability(available bool) {
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

func (entity *Number) ReportState(state int) {
	log.Printf("[%s] publishing state to: %s / %d", entity.ObjectId, entity.StateTopic, state)

	pubState := mqtt.C.Publish(entity.StateTopic, 0, false, strconv.Itoa(state))
	if pubState.Error() != nil {
		log.Fatalf("[%s] failed to publish state data: %s", entity.ObjectId, pubState.Error())
	}
}

func (entity *Number) SetValue(value int) error {

	return entity.valueSetter(value)
}
