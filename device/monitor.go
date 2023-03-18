package device

import (
	"ddmqtt/ddmrpc"
	"ddmqtt/structs"
	"log"
)

var Monitor structs.DiscoveryDevice

func InitMontor() structs.DiscoveryDevice {
	if Monitor != (structs.DiscoveryDevice{}) {
		return Monitor
	}
	attrs, err := ddmrpc.GetAssetAttributes()
	if err != nil {

		log.Fatalf("failed to read monitor info: %s", err.Error())
	}

	Monitor = structs.DiscoveryDevice{
		Identifiers:  attrs.ServiceTag,
		Manufacturer: "Dell",
		Model:        attrs.ModelCode,
		Name:         attrs.Model,
	}

	return Monitor
}

func GetMonitorActiveHours() (int, error) {
	hours, err := ddmrpc.GetMonitorActiveHours()

	if err != nil {
		return -1, err
	}

	return hours, nil
}