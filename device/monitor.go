package device

import (
	"ddmqtt/ddmrpc"
)

func GetAssetAttributes() (ddmrpc.AssetAttributes, error) {

	return ddmrpc.GetAssetAttributes()
}

func GetBrightnessLevel() (int, error) {

	return ddmrpc.GetBrightnessLevel()
}

func GetMonitorActiveHours() (int, error) {

	return ddmrpc.GetMonitorActiveHours()
}
