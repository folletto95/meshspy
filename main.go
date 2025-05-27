package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "github.com/tarm/serial"
    "google.golang.org/protobuf/proto"

    // i binding generati in pb/meshtastic
    pb "github.com/nicbad/meshspy/pb/meshtastic"
)

func leggiVarintFrame(r io.Reader) ([]byte, error) {
    var length uint64
    for shift := uint(0); ; shift += 7 {
        var b [1]byte
        if _, err := r.Read(b[:]); err != nil {
            return nil, err
        }
        length |= uint64(b[0]&0x7F) << shift
        if b[0]&0x80 == 0 {
            break
        }
    }
    frame := make([]byte, length)
    if _, err := io.ReadFull(r, frame); err != nil {
        return nil, err
    }
    return frame, nil
}

func main() {
    serialPort := getEnv("SERIAL_PORT", "/dev/ttyUSB0")
    baudRate := getEnvInt("BAUD_RATE", 115200)
    mqttBroker := getEnv("MQTT_BROKER", "tcp://mqtt-broker:1883")
    mqttTopic := getEnv("MQTT_TOPIC", "meshspy/nodo/connesso")
    clientID := getEnv("MQTT_CLIENT_ID", "meshspy-client")
    mqttUser := getEnv("MQTT_USER", "")
    mqttPass := getEnv("MQTT_PASS", "")

    cfg := &serial.Config{Name: serialPort, Baud: baudRate, ReadTimeout: time.Second * 5}
    port, err := serial.OpenPort(cfg)
    if err != nil {
        log.Fatalf("Impossibile aprire %s: %v", serialPort, err)
    }
    defer port.Close()

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

    log.Printf("In ascolto su %s a %d baud", serialPort, baudRate)
    reader := bufio.NewReader(port)

    for {
        frame, err := leggiVarintFrame(reader)
        if err != nil {
            if err == io.EOF {
                continue
            }
            log.Printf("Errore leggendo frame: %v", err)
            time.Sleep(time.Second)
            continue
        }

        // Unmarshal ServiceEnvelope
        var env pb.ServiceEnvelope
        if err := proto.Unmarshal(frame, &env); err != nil {
            log.Printf("Unmarshal envelope: %v", err)
            continue
        }
        // Unmarshal MeshPacket
        var pkt pb.MeshPacket
        if err := proto.Unmarshal(env.GetPayload(), &pkt); err != nil {
            log.Printf("Unmarshal packet: %v", err)
            continue
        }

        nodeID := fmt.Sprintf("0x%08x", pkt.GetFrom())
        text := ""
        if d := pkt.GetDecoded(); d != nil {
            text = d.GetText()
        }

        payload := fmt.Sprintf(`{"node":"%s","ts":%d,"text":"%s"}`,
            nodeID, time.Now().Unix(), text)
        tok := client.Publish(mqttTopic, 0, false, payload)
        tok.Wait()
        if err := tok.Error(); err != nil {
            log.Printf("Errore publish MQTT: %v", err)
        } else {
            log.Printf("Pubblicato: %s", payload)
        }
    }
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
