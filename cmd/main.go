//go:build windows

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/monitor"
	"ddmqtt/internal/mqtt"
	"ddmqtt/internal/registry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	go func() {
		<-ctx.Done()
		<-time.After(10 * time.Second)
		log.Fatal("service has not been stopped within the specified timeout; killed by force")
	}()

	cfg := config.InitConfig("config.json")

	registryClient := registry.NewRegistry(cfg)
	ddmRpcClient := ddmrpc.NewDdmRpc(registryClient)

	mqttClient := mqtt.NewClient(cfg)
	mqttErr := mqttClient.Connect()
	if mqttErr != nil {
		log.Fatalf("mqtt client connect: %s", mqttErr)
	}

	con := monitor.NewController(cfg, ddmRpcClient, mqttClient)
	errMonitors := con.StartReporting(ctx)
	if errMonitors != nil {
		log.Fatalf("monitors controller: %s", errMonitors)
	}
}
