package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/tarm/serial"
)

func main() {
    // Leggi configurazione da variabili d'ambiente
    serialPort := getEnv("SERIAL_PORT", "/dev/ttyUSB0")
    baudRate := getEnvInt("BAUD_RATE", 115200)
    mqttBroker := getEnv("MQTT_BROKER", "tcp://smpisa.ddns.net:1883")
    mqttTopic := getEnv("MQTT_TOPIC", "meshspy/nodo/connesso")
    clientID := getEnv("MQTT_CLIENT_ID", "meshspy-client")
	mqttUser   := getEnv("MQTT_USER", "testmeshspy")
	mqttPass   := getEnv("MQTT_PASS", "test1")

    // Apri la porta seriale
    cfg := &serial.Config{
        Name:        serialPort,
        Baud:        baudRate,
        ReadTimeout: 5 * time.Second,
    }
    port, err := serial.OpenPort(cfg)
    if err != nil {
        log.Fatalf("Impossibile aprire %s: %v", serialPort, err)
    }
    defer port.Close()

    // Connetti a MQTT
    opts := mqtt.NewClientOptions().
        AddBroker(mqttBroker).
        SetClientID(clientID)
	if mqttUser != "" {
		opts.SetUsername(mqttUser)
		opts.SetPassword(mqttPass)
	}
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("Connessione MQTT fallita: %v", token.Error())
    }
    defer client.Disconnect(250)

    log.Printf("In ascolto su seriale %s a %d baud", serialPort, baudRate)
    reader := bufio.NewReader(port)

    for {
        // Leggi una linea dalla seriale
        line, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("Errore lettura seriale: %v", err)
            time.Sleep(1 * time.Second)
            continue
        }

        nodeName := parseNodeName(line)
        if nodeName == "" {
            // riga non valida: ignora
            continue
        }

        // Crea il payload JSON
        payload := fmt.Sprintf(`{"node":"%s","ts":%d}`, nodeName, time.Now().Unix())

        // Pubblica su MQTT
        token := client.Publish(mqttTopic, 0, false, payload)
        token.Wait()
        if token.Error() != nil {
            log.Printf("Errore publish MQTT: %v", token.Error())
        } else {
            log.Printf("Pubblicato su %s: %s", mqttTopic, payload)
        }
    }
}

// parseNodeName estrae il nome del nodo dalla linea letta.
// Modifica il pattern secondo il formato effettivo dei dati.
func parseNodeName(line string) string {
    // Esempio: "NODE: berry001\n"
    var name string
    if n, _ := fmt.Sscanf(line, "NODE: %s\n", &name); n == 1 {
        return name
    }
    return ""
}

// getEnv restituisce il valore di KEY o def se non è settata
func getEnv(key, def string) string {
    if v, ok := os.LookupEnv(key); ok {
        return v
    }
    return def
}

// getEnvInt restituisce il valore numerico di KEY o def se non è settata o non valida
func getEnvInt(key string, def int) int {
    if v, ok := os.LookupEnv(key); ok {
        var i int
        if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
            return i
        }
    }
    return def
}
