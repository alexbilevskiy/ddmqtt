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

func ExecuteRaw(cmd string) (string, error) {

	return registry.ExecuteCommand(cmd)
}
