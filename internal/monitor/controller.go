package monitor

import (
	"context"
	"log"
	"time"

	"ddmqtt/internal/config"
	"ddmqtt/internal/mqtt"
)

type MonitorsController struct {
	cfg      *config.Config
	ddmrpc   DDMRPCClient
	mqtt     *mqtt.Client
	monitors map[string]*monitor
}

type monitor struct {
	cancel  context.CancelFunc
	monitor *Monitor
}

func NewController(cfg *config.Config, ddmrpc DDMRPCClient, mqtt *mqtt.Client) *MonitorsController {
	return &MonitorsController{
		cfg:      cfg,
		ddmrpc:   ddmrpc,
		mqtt:     mqtt,
		monitors: make(map[string]*monitor),
	}
}

func (c *MonitorsController) StartReporting(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			monitorsCount, err := c.ddmrpc.CountMonitors()
			if err != nil {
				log.Printf("failed to count monitors: %s", err)
				time.Sleep(1 * time.Second)
				continue
			}
			if monitorsCount == 0 {
				log.Printf("no monitors found")
				time.Sleep(1 * time.Second)
				continue
			}
			//log.Printf("%d monitors found", monitorsCount)
			// check if all known monitors are alive
			for tag, m := range c.monitors {
				_, errAttrs := c.ddmrpc.GetAssetAttributesByTag(tag)
				if errAttrs != nil {
					log.Printf("get asset attributes for monitor %s: %s", tag, errAttrs)
					m.cancel()
					delete(c.monitors, tag)
				}
			}
			if len(c.monitors) == monitorsCount {
				// nothing changed
				time.Sleep(1 * time.Second)
				continue
			}
			// something changed
			systemMonitors := make(map[string]struct{})
			for i := 0; i < monitorsCount; i++ {
				//log.Printf("describe monitor %d", i)
				attrs, errAttrs := c.ddmrpc.GetAssetAttributes(i + 1)
				if errAttrs != nil {
					//log.Printf("get asset attributes for monitor %d: %s", i, errAttrs)
				}
				systemMonitors[attrs.ServiceTag] = struct{}{}
				if _, ok := c.monitors[attrs.ServiceTag]; ok {
					//log.Printf("found existing monitor %s", attrs.ServiceTag)
					continue
				}
				log.Printf("found new monitor: %v", attrs)
				monCtx, cancel := context.WithCancel(ctx)
				c.monitors[attrs.ServiceTag] = &monitor{monitor: newMonitor(c.cfg, c.ddmrpc, c.mqtt, &attrs), cancel: cancel}
				go c.monitors[attrs.ServiceTag].monitor.StartReporting(monCtx)
			}
			for tag, _ := range c.monitors {
				if _, ok := systemMonitors[tag]; !ok {
					log.Printf("disconnected monitor: %v", tag)
					c.monitors[tag].cancel()
					delete(c.monitors, tag)
					continue
				}
				if _, ok := c.monitors[tag]; ok {
					continue
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}
