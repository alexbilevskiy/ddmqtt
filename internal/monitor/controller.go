package monitor

import (
	"fmt"
	"log"

	"ddmqtt/internal/config"
	"ddmqtt/internal/mqtt"
)

type MonitorsController struct {
	cfg    *config.Config
	ddmrpc DDMRPCClient
	mqtt   *mqtt.Client
}

func NewController(cfg *config.Config, ddmrpc DDMRPCClient, mqtt *mqtt.Client) *MonitorsController {
	return &MonitorsController{
		cfg:    cfg,
		ddmrpc: ddmrpc,
		mqtt:   mqtt,
	}
}

func (c *MonitorsController) PopulateMonitors() ([]*Monitor, error) {
	monitorsCount, err := c.ddmrpc.CountMonitors()
	if err != nil {
		return nil, fmt.Errorf("count monitors: %w", err)
	}
	if monitorsCount == 0 {
		return nil, fmt.Errorf("no monitors found")
	}
	monitors := make([]*Monitor, monitorsCount, monitorsCount)
	for i := 0; i < monitorsCount; i++ {
		attrs, err := c.ddmrpc.GetAssetAttributes(i + 1)
		if err != nil {
			return nil, fmt.Errorf("get asset attributes for monitor %d: %w", i, err)
		}
		log.Printf("found monitor: %v", attrs)
		monitors[i] = newMonitor(c.cfg, c.ddmrpc, c.mqtt, &attrs)
	}

	return monitors, nil
}
