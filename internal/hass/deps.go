package hass

type mqttClient interface {
	Publish(topic string, retained bool, payload interface{}) error
	AddListener(topic string, listener func(payload []byte))
}

type SensorReader func(identifier string) (int, error)

type NumberReader func(identifier string) (int, error)
type NumberSetter func(identifier string, value int) error

type SwitchReader func(identifier string) (string, error)
type SwitchSetter func(identifier string, value string) error

type ButtonSetter func(identifier string) error
