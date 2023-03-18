package mqtt

import (
	"ddmqtt/config"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var C mqtt.Client

func InitMqtt() {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(config.CFG.BrokerAddr).SetClientID(config.CFG.MqttClientId)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	C = mqtt.NewClient(opts)
	if token := C.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect: %s", token.Error())
	}
}
