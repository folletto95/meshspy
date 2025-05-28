package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "plugin"
    "regexp"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/tarm/serial"
)

var downloadFunc func(owner, repo, path, ref, out, token string) error

func init() {
    // Determina il percorso del .so rispetto all'eseguibile
    exePath, err := os.Executable()
    if err != nil {
        log.Fatalf("os.Executable fallito: %v", err)
    }
    exeDir := filepath.Dir(exePath)
    pluginPath := filepath.Join(exeDir, "ghdownloader.so")

    // Carica il plugin
    p, err := plugin.Open(pluginPath)
    if err != nil {
        log.Fatalf("plugin.Open fallito: %v", err)
    }
    sym, err := p.Lookup("DownloadProtos")
    if err != nil {
        log.Fatalf("Lookup DownloadProtos fallito: %v", err)
    }
    var ok bool
    downloadFunc, ok = sym.(func(string, string, string, string, string, string) error)
    if !ok {
        log.Fatalf("Firma di DownloadProtos non corrisponde")
    }
}

var nodeRe = regexp.MustCompile(`from=(0x[0-9a-fA-F]+)`)

func main() {
    // Configurazione da env
    serialPort := getEnv("SERIAL_PORT", "/dev/ttyUSB0")
    baudRate := getEnvInt("BAUD_RATE", 115200)
    mqttBroker := getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883")
    mqttTopic := getEnv("MQTT_TOPIC", "meshspy/nodo/connesso")
    clientID := getEnv("MQTT_CLIENT_ID", "meshspy-client")
    mqttUser := getEnv("MQTT_USER", "")
    mqttPass := getEnv("MQTT_PASS", "")

    // Setup seriale
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

    // Setup MQTT
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

    // **Scarica i .proto** all'avvio con il plugin
    if err := downloadFunc(
        "meshtastic",         // owner GitHub
        "protobufs",          // repo
        "meshtastic",         // path nella repo
        "v2.0.14",            // tag/branch
        "./meshtastic-proto", // cartella di destinazione
        os.Getenv("GH_TOKEN"),// token opzionale
    ); err != nil {
        log.Printf("Errore in DownloadProtos: %v", err)
    } else {
        log.Println("Plugin: download .proto completato")
    }

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
        if node == "" || node == "0x0" || node == lastNode {
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
        return m[1]
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
