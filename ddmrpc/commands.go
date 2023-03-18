package ddmrpc

import (
	"ddmqtt/registry"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const ResponseOk = "Ok"

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

func GetFirmwareVersion() (string, error) {
	res, err := registry.ExecuteCommand("GetFirmwareVersion")
	if err != nil {

		return "", err
	}

	return res, nil
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

	return strconv.Atoi(res)
}

func SetBrightnessLevel(brightness int) error {
	res, err := registry.ExecuteCommand("SetBrightnessLevel", strconv.Itoa(brightness))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetBrightnessLevel response: %s", res))
	}

	return nil
}

func GetContrastLevel() (int, error) {
	res, err := registry.ExecuteCommand("GetContrastLevel")
	if err != nil {

		return -1, err
	}

	return strconv.Atoi(res)
}

func SetContrastLevel(contrast int) error {
	res, err := registry.ExecuteCommand("SetContrastLevel", strconv.Itoa(contrast))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetContrastLevel response: %s", res))
	}

	return nil
}

func ExecuteRaw(cmd string) (string, error) {

	return registry.ExecuteCommand(cmd)
}
