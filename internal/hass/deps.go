package hass

type mqttClient interface {
	Publish(topic string, retained bool, payload interface{}) error
	AddListener(topic string, listener func(payload []byte))
}
