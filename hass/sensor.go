package hass

import (
	"ddmqtt/mqtt"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

func (entity *Sensor) GetType() string {
	return TypeSensor
}

func (entity *Sensor) SetValueReader(reader func() (int, error)) {
	entity.valueReader = reader
}

func (entity *Sensor) Init() error {
	entity.DoDiscovery()

	return nil
}

func (entity *Sensor) DoDiscovery() {
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

func (entity *Sensor) ReportValue() error {
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

func (entity *Sensor) reportAvailability(available bool) {
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

func (entity *Sensor) publishState(state int) {
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
