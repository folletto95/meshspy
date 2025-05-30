package main

import (
	"meshspy/config"
	"meshspy/mqtt"
	"meshspy/serial"
)

func main() {
	// Carica la configurazione dal file/environment
	cfg := config.Load()

	// Inizializza il client MQTT e si connette al broker
	client := mqtt.NewClient()
	client.Connect()

	// Crea un publisher MQTT con i parametri configurati
	pub := mqtt.NewPublisher(
		cfg.MQTTBroker,
		cfg.ClientID,
		cfg.MQTTTopic,
		cfg.User,
		cfg.Password,
		cfg.Debug,
	)
	defer pub.Close()

	// Avvia il loop di lettura dalla seriale e pubblica i dati via MQTT
	serial.ReadLoop(cfg.SerialPort, cfg.BaudRate, cfg.Debug, pub.Publish)
}
