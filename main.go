package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "regexp"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/tarm/serial"
)

// regex per catturare l'ID dopo "from="
var nodeRe = regexp.MustCompile(`from=(0x[0-9a-fA-F]+)`)

func main() {
    serialPort := getEnv("SERIAL_PORT", "/dev/ttyUSB0")
    baudRate := getEnvInt("BAUD_RATE", 115200)
    mqttBroker := getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883")
    mqttTopic := getEnv("MQTT_TOPIC", "meshspy/nodo/connesso")
    clientID := getEnv("MQTT_CLIENT_ID", "meshspy-client")
    mqttUser := getEnv("MQTT_USER", "")
    mqttPass := getEnv("MQTT_PASS", "")

    // setup serial
    cfg := &serial.Config{Name: serialPort, Baud: baudRate, ReadTimeout: time.Second * 5}
    port, err := serial.OpenPort(cfg)
    if err != nil {
        log.Fatalf("Impossibile aprire %s: %v", serialPort, err)
    }
    defer port.Close()

    // setup MQTT
    opts := mqtt.NewClientOptions().
        AddBroker(mqttBroker).
        SetClientID(clientID)
    if mqttUser != "" {
        opts.SetUsername(mqttUser)
        opts.SetPassword(mqttPass)
    }
    client := mqtt.NewClient(opts)
    if tok := client.Connect(); tok.Wait() && tok.Error() != nil {
        log.Fatalf("Connessione MQTT fallita: %v", tok.Error())
    }
    defer client.Disconnect(250)

    log.Printf("In ascolto su seriale %s a %d baud", serialPort, baudRate)
    reader := bufio.NewReader(port)

    var lastNode string

    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                continue
            }
            log.Printf("Errore lettura seriale: %v", err)
            time.Sleep(time.Second)
            continue
        }

        node := parseNodeName(line)
        // ignora righe senza ID, ID vuoto o "0x0"
        if node == "" || node == "0x0" {
            continue
        }
        // evita duplicati consecutivi
        if node == lastNode {
            continue
        }
        lastNode = node

        payload := fmt.Sprintf(`{"node":"%s","ts":%d}`, node, time.Now().Unix())
        tok := client.Publish(mqttTopic, 0, false, payload)
        tok.Wait()
        if tok.Error() != nil {
            log.Printf("Errore publish MQTT: %v", tok.Error())
        } else {
            log.Printf("Pubblicato su %s: %s", mqttTopic, payload)
        }
    }
}

func parseNodeName(line string) string {
    m := nodeRe.FindStringSubmatch(line)
    if len(m) == 2 {
        return m[1] // es. "0xbb210daf"
    }
    return ""
}

func getEnv(key, def string) string {
    if v, ok := os.LookupEnv(key); ok {
        return v
    }
    return def
}

func getEnvInt(key string, def int) int {
    if v, ok := os.LookupEnv(key); ok {
        var i int
        if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
            return i
        }
    }
    return def
}
