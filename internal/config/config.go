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
    MqttBroker string
    MqttTopic  string
}

func Load() Config {
    err := godotenv.Load(".env.runtime")
    if err != nil {
        log.Fatalf("Error loading .env.runtime file: %v", err)
    }

    baud, err := strconv.Atoi(os.Getenv("BAUD_RATE"))
    if err != nil {
        log.Fatalf("Invalid BAUD_RATE: %v", err)
    }

    return Config{
        SerialPort: os.Getenv("SERIAL_PORT"),
        BaudRate:   baud,
        MqttBroker: os.Getenv("MQTT_BROKER"),
        MqttTopic:  os.Getenv("MQTT_TOPIC"),
    }
}
