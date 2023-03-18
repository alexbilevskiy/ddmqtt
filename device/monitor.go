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

func SetBrightnessLevel(brightness int) error {

	return ddmrpc.SetBrightnessLevel(brightness)
}

func GetContrastLevel() (int, error) {

	return ddmrpc.GetContrastLevel()
}

func SetContrastLevel(contrast int) error {

	return ddmrpc.SetContrastLevel(contrast)
}

func GetMonitorActiveHours() (int, error) {

	return ddmrpc.GetMonitorActiveHours()
}
