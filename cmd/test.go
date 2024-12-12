//go:build windows

package main

import (
	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"ddmqtt/internal/registry"
	"log"
)

func main() {
	cfg := config.InitConfig("config.json")
	registryClient := registry.NewRegistry(cfg)
	ddm := ddmrpc.NewDdmRpc(registryClient)
	res, err := ddm.ExecuteRaw("SetBrightnessLevel 20")
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("result: %s", res)
}
