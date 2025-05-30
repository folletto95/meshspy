package mqtt

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"meshspy/config"
)

// Info rappresenta le informazioni estratte dal dispositivo Meshtastic
type Info struct {
	NodeName string
	Firmware string
}

// Espressioni regolari per estrarre nome del nodo e firmware
var nameRe = regexp.MustCompile(`long_name:"([^"]+)"`)
var fwRe = regexp.MustCompile(`FirmwareVersion\s+([^\s]+)`)

// GetInfo esegue meshtastic-go e recupera le informazioni dal dispositivo seriale
func GetInfo(port string) (*Info, error) {
	log.Println("📡 Invocazione meshtastic-go per la lettura informazioni...")

	cmd := exec.Command("meshtastic-go", "--port", port, "info")
	log.Printf("📤 Eseguo comando: %v", cmd.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Esecuzione del comando
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Errore durante l'esecuzione di meshtastic-go: %v\nOutput:\n%s", err, out.String())
		return nil, err
	}

	// Analisi dell'output riga per riga
	scanner := bufio.NewScanner(&out)
	info := &Info{}
	for scanner.Scan() {
		line := scanner.Text()
		if m := nameRe.FindStringSubmatch(line); len(m) == 2 {
			info.NodeName = strings.TrimSpace(m[1])
		}
		if m := fwRe.FindStringSubmatch(line); len(m) == 2 {
			info.Firmware = strings.TrimSpace(m[1])
		}
	}

	// Log finale per debug
	log.Printf("✅ Info trovate - Nodo: %s, Firmware: %s\n", info.NodeName, info.Firmware)
	log.Printf("ℹ️  Info dispositivo Meshtastic:\n%+v\n", info)
	return info, nil
}

// ConnectMQTT crea e restituisce un client MQTT connesso
func ConnectMQTT(cfg config.Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(cfg.ClientID)

	if cfg.User != "" {
		opts.SetUsername(cfg.User)
		opts.SetPassword(cfg.Password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	return client, token.Error()
}
