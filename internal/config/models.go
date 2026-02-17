package config

type Config struct {
	RegUser             string   `json:"registry_user"`
	BrokerAddr          string   `json:"broker_addr"`
	MqttClientId        string   `json:"mqtt_client_id"`
	MqttRootTopic       string   `json:"mqtt_root_topic"`
	HassDiscoveryPrefix string   `json:"hass_discovery_prefix"`
	Presets             []Preset `json:"presets"`
}

type Preset struct {
 	Name       string `json:"name"`
	Brightness int    `json:"brightness"`
	Contrast   int    `json:"contrast"`
}
