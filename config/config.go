// Package config loads application settings from environment variables.
package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"go.bug.st/serial/enumerator"
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
	MgmtURL      string
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

	serialPort := getEnv("SERIAL_PORT", "/dev/ttyUSB0")
	if !portExists(serialPort) {
		log.Printf("⚠️  porta seriale %s non trovata, ricerca automatica", serialPort)
		if p, err := autoDetectPort(); err == nil {
			serialPort = p
			log.Printf("✅ porta seriale %s selezionata", serialPort)
		} else {
			log.Printf("⚠️  nessuna porta seriale trovata: %v", err)
		}
	}

	return Config{
		SerialPort:   serialPort,
		BaudRate:     baud,
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883"),
		MQTTTopic:    getEnv("MQTT_TOPIC", "meshspy/nodo/connesso"),
		CommandTopic: getEnv("MQTT_COMMAND_TOPIC", "meshspy/commands"),
		ClientID:     getEnv("MQTT_CLIENT_ID", "meshspy-client"),
		User:         os.Getenv("MQTT_USER"),
		Password:     os.Getenv("MQTT_PASS"),
		Debug:        debug,
		SendAlive:    sendAlive,
		MgmtURL:      os.Getenv("MGMT_SERVER_URL"),
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func portExists(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func autoDetectPort() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}
	for _, p := range ports {
		if p.IsUSB {
			if portExists(p.Name) {
				return p.Name, nil
			}
		}
	}
	// fallback search by glob patterns
	patterns := []string{"/dev/ttyACM*", "/dev/ttyUSB*"}
	for _, pat := range patterns {
		if matches, _ := filepath.Glob(pat); len(matches) > 0 {
			if portExists(matches[0]) {
				return matches[0], nil
			}
		}
	}
	return "", errors.New("serial port not found")
}
