package ddmrpc

import (
	"ddmqtt/internal/config"
	"ddmqtt/internal/registry"
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

var mu sync.Mutex

type registryClient interface {
	ReadKey(direction string) (string, error)
	WriteCommand(command string, params ...string) error
}

type DdmRpc struct {
	cfg      *config.Config
	registry registryClient
}

func NewDdmRpc(cfg *config.Config) *DdmRpc {
	return &DdmRpc{
		cfg:      cfg,
		registry: registry.NewRegistry(cfg),
	}
}

func (d *DdmRpc) GetAssetAttributes() (AssetAttributes, error) {
	var asset AssetAttributes
	res, err := d.executeCommand("GetAssetAttributes", ReturnTypeString)
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

func (d *DdmRpc) GetFirmwareVersion() (string, error) {
	res, err := d.executeCommand("GetFirmwareVersion", ReturnTypeString)
	if err != nil {

		return "", err
	}

	return res, nil
}

func (d *DdmRpc) GetMonitorActiveHours() (int, error) {
	res, err := d.executeCommand("GetMonitorActiveHours", ReturnTypeInt)
	if err != nil {

		return -1, err
	}
	hours64, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(hours64), nil
}

func (d *DdmRpc) GetBrightnessLevel() (int, error) {
	res, err := d.executeCommand("GetBrightnessLevel", ReturnTypeInt)
	if err != nil {

		return -1, err
	}

	return strconv.Atoi(res)
}

func (d *DdmRpc) SetBrightnessLevel(brightness int) error {
	res, err := d.executeCommand("SetBrightnessLevel", ReturnTypeOk, strconv.Itoa(brightness))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetBrightnessLevel response: %s", res))
	}

	return nil
}

func (d *DdmRpc) GetContrastLevel() (int, error) {
	res, err := d.executeCommand("GetContrastLevel", ReturnTypeInt)
	if err != nil {

		return -1, err
	}

	return strconv.Atoi(res)
}

func (d *DdmRpc) SetContrastLevel(contrast int) error {
	res, err := d.executeCommand("SetContrastLevel", ReturnTypeOk, strconv.Itoa(contrast))
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetContrastLevel response: %s", res))
	}

	return nil
}

func (d *DdmRpc) GetPower() (string, error) {
	res, err := d.executeCommand("GetControl", ReturnTypeString, "D6")
	if err != nil {

		return "", err
	}
	if res == "0001" {

		return "ON", nil
	}
	if res == "0004" {

		return "OFF", nil
	}

	return "", errors.New(fmt.Sprintf("invalid GetPower response: %s", res))
}

func (d *DdmRpc) SetPower(value string) error {
	var arg string
	if value == "ON" {
		arg = "01"
	} else if value == "OFF" {
		arg = "04"
	} else {
		return errors.New("invalid power value to set")
	}
	res, err := d.executeCommand("SetControl", ReturnTypeOk, "D6", arg)
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid SetPower response: %s", res))
	}

	return nil
}

func (d *DdmRpc) Reset() error {
	res, err := d.executeCommand("ForceReset", ReturnTypeOk)
	if err != nil {

		return err
	}
	if res != ResponseOk {

		return errors.New(fmt.Sprintf("invalid ForceReset response: %s", res))
	}

	return nil
}

func (d *DdmRpc) executeCommand(command string, returnType string, params ...string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	err := d.registry.WriteCommand(command, params...)
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
		res, err := d.registry.ReadKey(registry.DirectionOut)
		if err != nil {

			return "", errors.New(fmt.Sprintf("execute error: %s", err.Error()))
		}
		if res == ResponseEmpty {
			if att > 3 {
				log.Printf("[%s] empty response, retrying (%d)", command, att)
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if res == ResponseWait {
			if att > 3 {
				log.Printf("[%s] waiting for response, retrying (%d)", command, att)
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

func (d *DdmRpc) ExecuteRaw(cmd string) (string, error) {

	return d.executeCommand(cmd, ReturnTypeString)
}
