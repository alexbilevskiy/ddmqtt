package ddmrpc

import (
	"ddmqtt/registry"
	"strconv"
	"strings"
)

func GetAssetAttributes() (AssetAttributes, error) {
	var asset AssetAttributes
	res, err := registry.ExecuteCommand("GetAssetAttributes")
	if err != nil {

		return asset, err
	}
	parts := strings.Split(res, ",")
	asset = AssetAttributes{
		ModelCode:    parts[0],
		Model:        parts[1],
		ServiceTag:   parts[2],
		Manufactured: parts[3],
	}
	hours, _ := strconv.ParseInt(parts[4], 10, 64)
	asset.ActiveHours = hours

	return asset, nil
}

func GetMonitorActiveHours() (int, error) {
	res, err := registry.ExecuteCommand("GetMonitorActiveHours")
	if err != nil {

		return -1, err
	}
	hours64, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(hours64), nil
}

func GetBrightnessLevel() (int, error) {
	res, err := registry.ExecuteCommand("GetBrightnessLevel")
	if err != nil {

		return -1, err
	}
	bri64, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(bri64), nil
}

func ExecuteRaw(cmd string) (string, error) {

	return registry.ExecuteCommand(cmd)
}
