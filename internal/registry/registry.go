package registry

import (
	"ddmqtt/internal/config"
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const BaseKey = `\Software\EnTech\RPC\`
const DirectionIn = `In`
const DirectionOut = `Out`

type Registry struct {
	cfg *config.Config
}

func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{cfg: cfg}
}

func (r *Registry) ReadKey(direction string) (string, error) {
	k, err := registry.OpenKey(registry.USERS, r.cfg.RegUser+BaseKey+direction, registry.QUERY_VALUE)
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

func (r *Registry) writeKey(direction string, value string) error {
	kw, err := registry.OpenKey(registry.USERS, r.cfg.RegUser+BaseKey+direction, registry.SET_VALUE)
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

func (r *Registry) deleteKey(direction string) error {
	kd, err := registry.OpenKey(registry.USERS, r.cfg.RegUser+BaseKey+direction, registry.SET_VALUE)
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

func (r *Registry) writeRandomI() error {
	kwi, err := registry.OpenKey(registry.USERS, r.cfg.RegUser+BaseKey+DirectionIn, registry.SET_VALUE)
	if err != nil {

		return errors.New(fmt.Sprintf("open write i key error: %s", err.Error()))
	}
	defer kwi.Close()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	err = kwi.SetStringValue("Incoming", strconv.Itoa(rnd.Intn(50000)))
	if err != nil {
		return errors.New(fmt.Sprintf("write command i key error: %s", err.Error()))
	}

	return nil
}

func (r *Registry) WriteCommand(command string, params ...string) error {
	err := r.deleteKey(DirectionOut)
	if err != nil {
		return err
	}
	err = r.writeRandomI()
	if err != nil {
		return err
	}
	err = r.writeKey(DirectionIn, fmt.Sprintf("%s %s", command, strings.Join(params, " ")))
	if err != nil {
		return err
	}

	return nil
}
