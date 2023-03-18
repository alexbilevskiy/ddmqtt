package ddmrpc

import (
	"ddmqtt/registry"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ResponseOk = "Ok"
const ResponseInvalidCommand = "Invalid command"
const ResponseEmpty = ""
const ResponseWait = "..."

const ReturnTypeInt = "int"
const ReturnTypeString = "string"
const ReturnTypeOk = "ok"

var wg sync.WaitGroup

func GetAssetAttributes() (AssetAttributes, error) {
	var asset AssetAttributes
	res, err := executeCommand("GetAssetAttributes", ReturnTypeString)
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
	res, err := executeCommand("GetFirmwareVersion", ReturnTypeString)
	if err != nil {

		return "", err
	}

	return res, nil
}

func GetMonitorActiveHours() (int, error) {
	res, err := executeCommand("GetMonitorActiveHours", ReturnTypeInt)
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
	res, err := executeCommand("GetBrightnessLevel", ReturnTypeInt)
	if err != nil {

		return -1, err
	}

	return strconv.Atoi(res)
}

func SetBrightnessLevel(brightness int) error {
	res, err := executeCommand("SetBrightnessLevel", ReturnTypeOk, strconv.Itoa(brightness))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetBrightnessLevel response: %s", res))
	}

	return nil
}

func GetContrastLevel() (int, error) {
	res, err := executeCommand("GetContrastLevel", ReturnTypeInt)
	if err != nil {

		return -1, err
	}

	return strconv.Atoi(res)
}

func SetContrastLevel(contrast int) error {
	res, err := executeCommand("SetContrastLevel", ReturnTypeOk, strconv.Itoa(contrast))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetContrastLevel response: %s", res))
	}

	return nil
}

func executeCommand(command string, returnType string, params ...string) (string, error) {
	wg.Wait()
	wg.Add(1)
	defer wg.Done()

	err := registry.WriteCommand(command, params...)
	if err != nil {
		return "", errors.New(fmt.Sprintf("execute error: %s", err.Error()))
	}

	att := 0
	for {
	LOOP:
		att++
		if att > 10 {

			return "", errors.New(fmt.Sprintf("attempt limit reached for %s", command))
		}
		res, err := registry.ReadKey(registry.DirectionOut)
		if err != nil {

			return "", errors.New(fmt.Sprintf("execute error: %s", err.Error()))
		}
		if res == ResponseEmpty {
			if att > 3 {
				log.Printf("empty response, retrying (%d)", att)
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if res == ResponseWait {
			if att > 3 {
				log.Printf("waiting for response, retrying (%d)", att)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if res == ResponseInvalidCommand {

			return "", errors.New(fmt.Sprintf("execute error: %s", res))
		}
		switch returnType {
		case ReturnTypeString:
		case ReturnTypeInt:
			_, err = strconv.Atoi(res)
			if err != nil {
				if att > 3 {
					log.Printf("[%s] not int response, retrying (%d) `%s`: %s", command, att, res, err.Error())
				}
				time.Sleep(200 * time.Millisecond)
				goto LOOP
			}
		case ReturnTypeOk:
			{
				if res != ResponseOk {
					if att > 3 {
						log.Printf("[%s] not ok response, retrying (%d): %s", command, att, res)
					}
					time.Sleep(200 * time.Millisecond)
					goto LOOP
				}
			}
		}
		return res, nil
	}
}

func ExecuteRaw(cmd string) (string, error) {

	return executeCommand(cmd, ReturnTypeString)
}
