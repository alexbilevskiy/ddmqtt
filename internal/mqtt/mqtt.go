package mqtt

import (
	"fmt"
	"log"
	"os"
	"time"

	"ddmqtt/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	cfg       *config.Config
	client    mqtt.Client
	listeners map[string]func(client mqtt.Client, msg mqtt.Message)
}

func NewClient(cfg *config.Config) *Client {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	return &Client{
		cfg:       cfg,
		listeners: make(map[string]func(client mqtt.Client, msg mqtt.Message)),
	}
}

func (m *Client) Connect() error {
	opts := mqtt.NewClientOptions().AddBroker(m.cfg.BrokerAddr).SetClientID(m.cfg.MqttClientId)
	opts.SetKeepAlive(5 * time.Second)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MQTT message received from: %s\n", msg.Topic())
	})
	opts.SetPingTimeout(3 * time.Second)
	opts.WillEnabled = true
	opts.WillTopic = fmt.Sprintf("%s/available", m.cfg.MqttRootTopic)
	opts.WillPayload = []byte("offline")
	opts.WillRetained = true
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("MQTT connected!")
		if len(m.listeners) > 0 {
			m.client.Publish(opts.WillTopic, 0, true, "online")
		}
		for topic, listener := range m.listeners {
			if token := m.client.Subscribe(topic, 0, listener); token.Wait() && token.Error() != nil {

				log.Fatalf("failed to re-subscribe to %s: %s", topic, token.Error())
			}
		}
	})
	opts.SetAutoReconnect(true)

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect: %s", token.Error())
	}

	return nil
}

func (m *Client) AddListener(topic string, listener func(payload []byte)) {
	m.listeners[topic] = func(client mqtt.Client, message mqtt.Message) {
		if message.Topic() != topic {
			return
		}
		listener(message.Payload())
	}
	if token := m.client.Subscribe(topic, 0, m.listeners[topic]); token.Wait() && token.Error() != nil {

		log.Printf("failed to subscribe to %s: %s", topic, token.Error())
	}
}

func (m *Client) Publish(topic string, retain bool, value interface{}) error {
	if token := m.client.Publish(topic, 0, retain, value); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to %s: %s", topic, token.Error())
	}
	return nil
}
