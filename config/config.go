package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SerialPort string
	BaudRate   int
	MQTTBroker string
	MQTTTopic  string
	ClientID   string
	User       string
	Password   string
	Debug      bool
}

func Load() Config {
	_ = godotenv.Load(".env.runtime")

	baud, err := strconv.Atoi(getEnv("BAUD_RATE", "115200"))
	if err != nil {
		log.Fatalf("Invalid BAUD_RATE: %v", err)
	}

	debug := getEnv("DEBUG", "false") == "true"

	return Config{
		SerialPort: getEnv("SERIAL_PORT", "/dev/ttyUSB0"),
		BaudRate:   baud,
		MQTTBroker: getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883"),
		MQTTTopic:  getEnv("MQTT_TOPIC", "meshspy/nodo/connesso"),
		ClientID:   getEnv("MQTT_CLIENT_ID", "meshspy-client"),
		User:       os.Getenv("MQTT_USER"),
		Password:   os.Getenv("MQTT_PASS"),
		Debug:      debug,
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
