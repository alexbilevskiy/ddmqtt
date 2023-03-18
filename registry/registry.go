package registry

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const BaseKey = `Software\EnTech\RPC\`
const DirectionIn = `In`
const DirectionOut = `Out`

func ReadKey(direction string) (string, error) {
	key := BaseKey + direction
	k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", errors.New(fmt.Sprintf("read command key error: %s", err.Error()))
	}
	defer k.Close()

	s, _, err := k.GetStringValue("")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", errors.New(fmt.Sprintf("read command error: %s", err))
	}

	return s, nil
}

func writeKey(direction string, value string) error {
	kw, err := registry.OpenKey(registry.CURRENT_USER, BaseKey+direction, registry.SET_VALUE)
	if err != nil {

		return errors.New(fmt.Sprintf("open write key error: %s", err.Error()))
	}
	defer kw.Close()
	err = kw.SetStringValue("", value)
	if err != nil {
		return errors.New(fmt.Sprintf("write command key error: %s", err.Error()))
	}

	return nil
}

func deleteKey(direction string) error {
	kd, err := registry.OpenKey(registry.CURRENT_USER, BaseKey+direction, registry.SET_VALUE)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return errors.New(fmt.Sprintf("open delete key error: %s", err.Error()))
	}
	defer kd.Close()

	err = kd.DeleteValue("")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return errors.New(fmt.Sprintf("delete key error: %s", err.Error()))
	}

	return nil
}

func writeRandomI() error {
	kwi, err := registry.OpenKey(registry.CURRENT_USER, BaseKey+DirectionIn, registry.SET_VALUE)
	if err != nil {

		return errors.New(fmt.Sprintf("open write i key error: %s", err.Error()))
	}
	defer kwi.Close()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	err = kwi.SetStringValue("Incoming", strconv.Itoa(r.Intn(50000)))
	if err != nil {
		return errors.New(fmt.Sprintf("write command i key error: %s", err.Error()))
	}

	return nil
}

func WriteCommand(command string, params ...string) error {
	deleteKey(DirectionOut)
	writeRandomI()
	writeKey(DirectionIn, fmt.Sprintf("%s %s", command, strings.Join(params, " ")))

	return nil
}
