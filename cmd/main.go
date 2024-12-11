//go:build windows

package main

import (
	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/monitor"
	"ddmqtt/internal/mqtt"
)

func main() {
	cfg := config.InitConfig("config.json")
	ddmRpcClient := ddmrpc.NewDdmRpc(cfg)
	mqttClient := mqtt.NewClient(cfg)
	mon := monitor.NewMonitor(cfg, ddmRpcClient, mqttClient)
	mon.StartReporting()
}
