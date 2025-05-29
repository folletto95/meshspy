package main

import (
	"meshspy/config"
	"meshspy/mqtt"
	"meshspy/serial"
)

func main() {
	cfg := config.Load()
	pub := mqtt.NewPublisher(cfg.MQTTBroker, cfg.ClientID, cfg.MQTTTopic, cfg.User, cfg.Password, cfg.Debug)
	defer pub.Close()

	serial.ReadLoop(cfg.SerialPort, cfg.BaudRate, cfg.Debug, pub.Publish)
}
