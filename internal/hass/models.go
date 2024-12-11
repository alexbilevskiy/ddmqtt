package hass

type Device struct {
	Identifiers  string `json:"identifiers"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Name         string `json:"name"`
	SwVersion    string `json:"sw_version"`
}

type SAvailability struct {
	Topic string `json:"topic"`
}
