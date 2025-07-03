// Package config loads application settings from environment variables.
package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	SerialPort   string
	BaudRate     int
	MQTTBroker   string
	MQTTTopic    string
	CommandTopic string
	ClientID     string
	User         string
	Password     string
	Debug        bool
	SendAlive    bool
}

// Load reads configuration values from the environment and returns a Config.
func Load() Config {

	baud, err := strconv.Atoi(getEnv("BAUD_RATE", "115200"))
	if err != nil {
		log.Fatalf("Invalid BAUD_RATE: %v", err)
	}

	debug := getEnv("DEBUG", "false") == "true"
	sendAlive := getEnv("SEND_ALIVE_ON_START", "false") == "true"

	return Config{
		SerialPort:   getEnv("SERIAL_PORT", "/dev/ttyUSB0"),
		BaudRate:     baud,
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883"),
		MQTTTopic:    getEnv("MQTT_TOPIC", "meshspy/nodo/connesso"),
		CommandTopic: getEnv("MQTT_COMMAND_TOPIC", "meshspy/commands"),
		ClientID:     getEnv("MQTT_CLIENT_ID", "meshspy-client"),
		User:         os.Getenv("MQTT_USER"),
		Password:     os.Getenv("MQTT_PASS"),
		Debug:        debug,
		SendAlive:    sendAlive,
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
