package config

type Config struct {
	BrokerAddr          string `json:"broker_addr"`
	MqttClientId        string `json:"mqtt_client_id"`
	HassDiscoveryPrefix string `json:"hass_discovery_prefix"`
}
