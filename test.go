//go:build windows

package main

import (
	"ddmqtt/ddmrpc"
	"log"
)

func main() {
	res, err := ddmrpc.ExecuteRaw("SetBrightnessLevel 20")
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("result: %s", res)
}
