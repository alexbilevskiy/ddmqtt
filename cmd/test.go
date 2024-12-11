//go:build windows

package main

import (
	"ddmqtt/internal/config"
	"ddmqtt/internal/ddmrpc"
	"log"
)

func main() {
	cfg := config.InitConfig("config.json")
	ddm := ddmrpc.NewDdmRpc(cfg)
	res, err := ddm.ExecuteRaw("SetBrightnessLevel 20")
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("result: %s", res)
}
