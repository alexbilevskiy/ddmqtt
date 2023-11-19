package mqtt

import (
	"ddmqtt/config"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

var C mqtt.Client

var l map[string]func(client mqtt.Client, msg mqtt.Message)

func InitMqtt(monitorIdentifiers string) {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(config.CFG.BrokerAddr).SetClientID(config.CFG.MqttClientId)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MQTT message received from: %s\n", msg.Topic())
	})
	opts.SetPingTimeout(1 * time.Second)
	opts.WillEnabled = true
	opts.WillTopic = fmt.Sprintf("%s/%s/available", config.CFG.MqttRootTopic, monitorIdentifiers)
	opts.WillPayload = []byte("offline")
	opts.WillRetained = true
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("MQTT connected!")
		for topic, listener := range l {
			if token := C.Subscribe(topic, 0, listener); token.Wait() && token.Error() != nil {

				log.Fatalf("failed to re-subscribe to %s: %s", topic, token.Error())
			}
		}
	})
	opts.SetAutoReconnect(true)

	C = mqtt.NewClient(opts)
	if token := C.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect: %s", token.Error())
	}
	l = make(map[string]func(client mqtt.Client, msg mqtt.Message))
}

func AddListener(topic string, listener func(client mqtt.Client, msg mqtt.Message)) {
	l[topic] = listener
	if token := C.Subscribe(topic, 0, listener); token.Wait() && token.Error() != nil {

		log.Fatalf("failed to subscribe to %s: %s", topic, token.Error())
	}
}
