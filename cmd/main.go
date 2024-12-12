//go:build windows

package main

import (
	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/monitor"
	"ddmqtt/internal/mqtt"
	"ddmqtt/internal/registry"
)

func main() {
	cfg := config.InitConfig("config.json")

	registryClient := registry.NewRegistry(cfg)
	ddmRpcClient := ddmrpc.NewDdmRpc(registryClient)

	mqttClient := mqtt.NewClient(cfg)

	mon := monitor.NewMonitor(cfg, ddmRpcClient, mqttClient)
	mon.StartReporting()
}
