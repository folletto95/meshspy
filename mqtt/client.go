package mqtt

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Publisher struct {
	client mqtt.Client
	topic  string
	debug  bool
}

func NewPublisher(broker, clientID, topic, user, pass string, debug bool) *Publisher {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)
	if user != "" {
		opts.SetUsername(user)
		opts.SetPassword(pass)
	}
	client := mqtt.NewClient(opts)
	if tok := client.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatalf("MQTT connection failed: %v", tok.Error())
	}
	return &Publisher{client: client, topic: topic, debug: debug}
}

func (p *Publisher) Publish(payload string) {
	tok := p.client.Publish(p.topic, 0, false, payload)
	tok.Wait()
	if err := tok.Error(); err != nil {
		log.Printf("MQTT publish error: %v", err)
	} else if p.debug {
		log.Printf("Published: %s", payload)
	}
}

func (p *Publisher) Close() {
	p.client.Disconnect(250)
}
