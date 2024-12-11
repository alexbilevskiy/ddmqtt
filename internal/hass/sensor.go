package hass

import (
	"encoding/json"
	"log"
	"strconv"
)

type Sensor struct {
	Discovered     bool   `json:"-"`
	Avaialable     bool   `json:"-"`
	State          int    `json:"-"`
	BaseTopic      string `json:"-"`
	DiscoveryTopic string `json:"-"`
	valueReader    func() (int, error)
	mqtt           mqttClient
	Name           string        `json:"name"`
	Availability   SAvailability `json:"availability"`
	StateTopic     string        `json:"state_topic"`
	ObjectId       string        `json:"object_id"`
	UniqueId       string        `json:"unique_id"`
	Device         *Device       `json:"device"`
	Icon           string        `json:"icon"`
}

func (entity *Sensor) SetMqtt(mqtt mqttClient) {
	entity.mqtt = mqtt
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

	js, _ := json.Marshal(entity)
	log.Printf("[%s] publishing discovery: %s", entity.ObjectId, string(js))

	err := entity.mqtt.Publish(entity.DiscoveryTopic, true, js)
	if err != nil {
		log.Printf("[%s] failed to publish discovery: %s", entity.ObjectId, err.Error())
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

	err := entity.mqtt.Publish(entity.Availability.Topic, false, availabilityStatus)
	if err != nil {
		log.Printf("[%s] failed to publish online state: %s", entity.ObjectId, err.Error())
	}
	entity.Avaialable = available
}

func (entity *Sensor) publishState(state int) {
	if entity.State == state {
		return
	}
	log.Printf("[%s] publishing state: %d", entity.ObjectId, state)

	err := entity.mqtt.Publish(entity.StateTopic, false, strconv.Itoa(state))
	if err != nil {
		log.Printf("[%s] failed to publish state data: %s", entity.ObjectId, err.Error())
	}
	entity.State = state
}
