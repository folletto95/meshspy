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
	EnableGUI    bool
}

// Load reads configuration values from the environment and returns a Config.
func Load() Config {

	baud, err := strconv.Atoi(getEnv("BAUD_RATE", "115200"))
	if err != nil {
		log.Fatalf("Invalid BAUD_RATE: %v", err)
	}

	debugStr := getEnv("DEBUG", "false")
	debug, err := strconv.ParseBool(debugStr)
	if err != nil {
		log.Printf("invalid DEBUG value %q, defaulting to false", debugStr)
		debug = false
	}

	aliveStr := getEnv("SEND_ALIVE_ON_START", "false")
	sendAlive, err := strconv.ParseBool(aliveStr)
	if err != nil {
		log.Printf("invalid SEND_ALIVE_ON_START value %q, defaulting to false", aliveStr)
		sendAlive = false
	}

	guiStr := getEnv("ENABLE_GUI", "false")
	enableGUI, err := strconv.ParseBool(guiStr)
	if err != nil {
		log.Printf("invalid ENABLE_GUI value %q, defaulting to false", guiStr)
		enableGUI = false
	}

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
		EnableGUI:    enableGUI,
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
