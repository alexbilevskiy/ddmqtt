package registry

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"

	"ddmqtt/internal/config"
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
	path := r.buildRegistryPath(direction)
	k, err := registry.OpenKey(registry.USERS, path, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("open registry key for reading (%s): %w", path, err)
	}
	defer k.Close()

	s, _, err := k.GetStringValue("")
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("read registry key (%s): %w", path, err)
	}

	return s, nil
}

func (r *Registry) writeKey(direction string, value string) error {
	path := r.buildRegistryPath(direction)
	kw, err := registry.OpenKey(registry.USERS, path, registry.SET_VALUE)
	if err != nil {

		return fmt.Errorf("open registry key for writing (%s): %w", path, err)
	}
	defer kw.Close()
	err = kw.SetStringValue("", value)
	if err != nil {
		return fmt.Errorf("write registry key (%s): %w", path, err)
	}

	return nil
}

func (r *Registry) deleteKey(direction string) error {
	path := r.buildRegistryPath(direction)
	kd, err := registry.OpenKey(registry.USERS, path, registry.SET_VALUE)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return fmt.Errorf("open registry key for deletion (%s): %w", path, err)
	}
	defer kd.Close()

	err = kd.DeleteValue("")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return fmt.Errorf("delete registry key (%s): %w", path, err)
	}

	return nil
}

func (r *Registry) writeRandomI() error {
	path := r.buildRegistryPath(DirectionIn)
	kwi, err := registry.OpenKey(registry.USERS, path, registry.SET_VALUE)
	if err != nil {

		return fmt.Errorf("open registry key for random write (%s): %w", path, err)
	}
	defer kwi.Close()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	err = kwi.SetStringValue("Incoming", strconv.Itoa(rnd.Intn(50000)))
	if err != nil {
		return fmt.Errorf("write random i (%s): %w", path, err)
	}

	return nil
}

func (r *Registry) WriteCommand(command string, params ...string) error {
	err := r.deleteKey(DirectionOut)
	if err != nil {
		return fmt.Errorf("registry WriteCommand: %w", err)
	}
	err = r.writeRandomI()
	if err != nil {
		return fmt.Errorf("registry WriteCommand: %w", err)
	}
	err = r.writeKey(DirectionIn, fmt.Sprintf("%s %s", command, strings.Join(params, " ")))
	if err != nil {
		return fmt.Errorf("registry WriteCommand: %w", err)
	}

	return nil
}

func (r *Registry) buildRegistryPath(direction string) string{
	return r.cfg.RegUser + BaseKey + direction
}