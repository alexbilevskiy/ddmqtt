package structs

type Config struct {
	BrokerAddr          string `json:"broker_addr"`
	MqttClientId        string `json:"mqtt_client_id"`
	HassDiscoveryPrefix string `json:"hass_discovery_prefix"`
}

type DiscoveryDevice struct {
	Identifiers  string `json:"identifiers"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Name         string `json:"name"`
}

type DiscoverySensor struct {
	Name string `json:"name"`
	//DeviceClass       string          `json:"device_class"`
	//AvailabilityTopic string          `json:"availability_topic"`
	Availability SAvailability   `json:"availability"`
	StateTopic   string          `json:"state_topic"`
	ObjectId     string          `json:"object_id"`
	UniqueId     string          `json:"unique_id"`
	Device       DiscoveryDevice `json:"device"`
}

type SAvailability struct {
	Topic string `json:"topic"`
}
