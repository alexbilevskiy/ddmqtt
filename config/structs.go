package config

type Config struct {
	BrokerAddr          string   `json:"broker_addr"`
	MqttClientId        string   `json:"mqtt_client_id"`
	HassDiscoveryPrefix string   `json:"hass_discovery_prefix"`
	Presets             []Preset `json:"presets"`
}

type Preset struct {
	Name       string
	Brightness int
	Contrast   int
}
