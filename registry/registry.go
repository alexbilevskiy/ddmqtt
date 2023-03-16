package registry

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func ReadKey(direction string) (string, error) {
	key := `Software\EnTech\RPC\` + direction
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

func WriteCommand(command string, params ...string) error {
	kd, err := registry.OpenKey(registry.CURRENT_USER, `Software\EnTech\RPC\Out`, registry.SET_VALUE)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return errors.New(fmt.Sprintf("open delete key error: %s", err.Error()))
	}
	defer kd.Close()

	err = kd.DeleteValue("")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {

		return errors.New(fmt.Sprintf("delete key error: %s", err.Error()))
	}

	kw, err := registry.OpenKey(registry.CURRENT_USER, `Software\EnTech\RPC\In`, registry.SET_VALUE)
	if err != nil {

		return errors.New(fmt.Sprintf("open write key error: %s", err.Error()))
	}
	defer kw.Close()
	err = kw.SetStringValue("", fmt.Sprintf("%s %s", command, strings.Join(params, " ")))
	if err != nil {
		return errors.New(fmt.Sprintf("write command key error: %s", err.Error()))
	}

	kwi, err := registry.OpenKey(registry.CURRENT_USER, `Software\EnTech\RPC\In`, registry.SET_VALUE)
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

func ExecuteCommand(command string, params ...string) (string, error) {
	err := WriteCommand(command, params...)
	if err != nil {
		return "", errors.New(fmt.Sprintf("execute error: %s", err.Error()))
	}

	att := 0
	for {
		att++
		res, err := ReadKey("Out")
		if err != nil {

			return "", errors.New(fmt.Sprintf("execute error: %s", err.Error()))
		}
		if res == "" {
			if att > 3 {
				log.Printf("empty response, retrying (%d)", att)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if res == "..." {
			if att > 3 {
				log.Printf("empty2 response, retrying (%d)", att)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return res, nil
	}
}
