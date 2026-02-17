package ddmrpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"ddmqtt/internal/registry"
)

const (
	ResponseOk             = "Ok"
	ResponseInvalidCommand = "Invalid command"
	ResponseEmpty          = ""
	ResponseWait           = "..."
)

const (
	ReturnTypeInt    = "int"
	ReturnTypeString = "string"
	ReturnTypeOk     = "ok"
)

var (
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidArg      = errors.New("invalid argument")
	ErrExecuteError    = errors.New("execute error")
)

type registryClient interface {
	ReadKey(direction string) (string, error)
	WriteCommand(command string, params ...string) error
}

type DdmRpc struct {
	mu       sync.Mutex
	registry registryClient
}

func NewDdmRpc(registry registryClient) *DdmRpc {
	return &DdmRpc{
		mu:       sync.Mutex{},
		registry: registry,
	}
}

func (d *DdmRpc) CountMonitors() (int, error) {
	res, err := d.executeCommand("CountMonitors", ReturnTypeInt)
	if err != nil {

		return -1, err
	}
	countMonitors, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return -1, fmt.Errorf("CountMonitors: %w", err)
	}

	return int(countMonitors), nil
}

func (d *DdmRpc) GetAssetAttributes(monitorId int) (AssetAttributes, error) {
	var asset AssetAttributes
	res, err := d.executeCommand(fmt.Sprintf("%d:GetAssetAttributes", monitorId), ReturnTypeString)
	if err != nil {

		return asset, fmt.Errorf("GetAssetAttributes: %w", err)
	}
	parts := strings.Split(res, ",")
	if parts[0] == "" {
		return asset, fmt.Errorf("GetAssetAttributes: %w", ErrInvalidResponse)
	}
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

func (d *DdmRpc) GetCapabilities(serviceTag string) (Capabilities, error) {
	var caps Capabilities
	res, err := d.executeCommand(fmt.Sprintf("%s:GetCapabilities", serviceTag), ReturnTypeString)
	if err != nil {

		return caps, fmt.Errorf("GetCapabilities: %w", err)
	}
	re := regexp.MustCompile(CapabilitiesRegex)
	matches := re.FindAllStringSubmatch(res, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return caps, fmt.Errorf("GetCapabilities: %w (%s)", ErrInvalidResponse, res)
	}
	for _, input := range strings.Split(strings.Trim(matches[0][1], " "), " ") {
		in, err := hex.DecodeString(input)
		if err != nil {
			log.Printf("invalid input code: %s", input)
			continue
		}
		caps.AvailableInputs = append(caps.AvailableInputs, in[0])
	}

	return caps, nil
}

func (d *DdmRpc) GetFirmwareVersion(serviceTag string) (string, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetFirmwareVersion", serviceTag), ReturnTypeString)
	if err != nil {

		return "", fmt.Errorf("GetFirmwareVersion: %w", err)
	}

	return res, nil
}

func (d *DdmRpc) GetMonitorActiveHours(serviceTag string) (int, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetMonitorActiveHours", serviceTag), ReturnTypeInt)
	if err != nil {

		return -1, fmt.Errorf("GetMonitorActiveHours: %w", err)
	}
	hours64, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return -1, fmt.Errorf("GetMonitorActiveHours: parse response: %w", err)
	}

	return int(hours64), nil
}

func (d *DdmRpc) GetBrightnessLevel(serviceTag string) (int, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetBrightnessLevel", serviceTag), ReturnTypeInt)
	if err != nil {

		return -1, fmt.Errorf("GetBrightnessLevel: %w", err)
	}

	return strconv.Atoi(res)
}

func (d *DdmRpc) SetBrightnessLevel(serviceTag string, brightness int) error {
	res, err := d.executeCommand(fmt.Sprintf("%s:SetBrightnessLevel", serviceTag), ReturnTypeOk, strconv.Itoa(brightness))
	if err != nil {

		return fmt.Errorf("SetBrightnessLevel: %w", err)
	}
	if res != ResponseOk {

		return fmt.Errorf("SetBrightnessLevel: %w (%s)", ErrInvalidResponse, res)
	}

	return nil
}

func (d *DdmRpc) GetContrastLevel(serviceTag string) (int, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetContrastLevel", serviceTag), ReturnTypeInt)
	if err != nil {

		return -1, fmt.Errorf("GetContrastLevel: %w", err)
	}
	level, err := strconv.Atoi(res)
	if err != nil {
		return -1, fmt.Errorf("GetContrastLevel: parse response: %w", err)
	}

	return level, nil
}

func (d *DdmRpc) SetContrastLevel(serviceTag string, contrast int) error {
	res, err := d.executeCommand(fmt.Sprintf("%s:SetContrastLevel", serviceTag), ReturnTypeOk, strconv.Itoa(contrast))
	if err != nil {

		return fmt.Errorf("SetContrastLevel: %w", err)
	}
	if res != ResponseOk {

		return fmt.Errorf("SetContrastLevel: %w (%s)", ErrInvalidResponse, res)
	}

	return nil
}

func (d *DdmRpc) GetPower(serviceTag string) (string, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetControl", serviceTag), ReturnTypeString, "D6")
	if err != nil {

		return "", fmt.Errorf("GetPower: %w", err)
	}
	if res == "0001" {

		return "ON", nil
	}
	if res == "0004" {

		return "OFF", nil
	}

	return "", fmt.Errorf("GetPower: %w (%s)", ErrInvalidResponse, res)
}

func (d *DdmRpc) SetPower(serviceTag string, value string) error {
	var arg string
	if value == "ON" {
		arg = "01"
	} else if value == "OFF" {
		arg = "04"
	} else {
		return fmt.Errorf("SetPower: %w", ErrInvalidArg)
	}
	res, err := d.executeCommand(fmt.Sprintf("%s:SetControl", serviceTag), ReturnTypeOk, "D6", arg)
	if err != nil {

		return fmt.Errorf("SetPower: %w", err)
	}
	if res != ResponseOk {

		return fmt.Errorf("SetPower: %w (%s)", ErrInvalidResponse, res)
	}

	return nil
}

func (d *DdmRpc) GetActiveInput(serviceTag string) (byte, error) {
	res, err := d.executeCommand(fmt.Sprintf("%s:GetControl 60", serviceTag), ReturnTypeString)
	if err != nil {

		return 0, fmt.Errorf("GetActiveInput: %w", err)
	}
	if len(res) != 4 {

		return 0, fmt.Errorf("GetActiveInput: %w (%s)", ErrInvalidResponse, res)
	}
	in, err := hex.DecodeString(res)
	if err != nil {

		return 0, fmt.Errorf("GetActiveInput: decode: %w", err)
	}
	return in[1], nil
}

func (d *DdmRpc) SetActiveInput(serviceTag string, input byte) error {
	res, err := d.executeCommand(fmt.Sprintf("%s:SetActiveInput", serviceTag), ReturnTypeOk, fmt.Sprintf("%02x", input))
	if err != nil {

		return fmt.Errorf("SetActiveInput: %w", err)
	}
	if res != ResponseOk {

		return fmt.Errorf("SetActiveInput: %w (%s)", ErrInvalidResponse, res)
	}

	return nil
}

func (d *DdmRpc) Reset(serviceTag string) error {
	res, err := d.executeCommand(fmt.Sprintf("%s:ForceReset", serviceTag), ReturnTypeOk)
	if err != nil {

		return fmt.Errorf("ForceReset: %w", err)
	}
	if res != ResponseOk {

		return fmt.Errorf("ForceReset: %w (%s)", ErrInvalidResponse, res)
	}

	return nil
}

func (d *DdmRpc) executeCommand(command string, returnType string, params ...string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.registry.WriteCommand(command, params...)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrExecuteError, err)
	}

	att := 0
	for {
	LOOP:
		att++
		if att > 10 {

			return "", fmt.Errorf("%w: attempt limit reached for %s", ErrExecuteError, command)
		}
		res, err := d.registry.ReadKey(registry.DirectionOut)
		if err != nil {

			return "", fmt.Errorf("%w: %w", ErrExecuteError, err)
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

			return "", fmt.Errorf("%w: invalid command", ErrExecuteError)
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
