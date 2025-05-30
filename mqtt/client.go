package mqtt

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"meshspy/config"
)

// Info rappresenta le informazioni estratte dal dispositivo Meshtastic
type Info struct {
	NodeName string
	Firmware string
}

// struttura per decodificare il JSON restituito da meshtastic-go --json info
type meshInfo struct {
	MyNodeNum       string                 `json:"my_node_num"`
	User            map[string]struct {
		LongName string `json:"long_name"`
	} `json:"user"`
	DeviceMetrics struct {
		FirmwareVersion string `json:"firmware_version"`
	} `json:"device_metrics"`
}

// GetInfo esegue meshtastic-go con output JSON e recupera le informazioni dal nodo locale
func GetInfo(port string) (*Info, error) {
	log.Println("📡 Invocazione meshtastic-go per la lettura informazioni.")
	cmd := exec.Command("meshtastic-go", "--port", port, "info", "--json")
	log.Printf("📤 Eseguo comando: %v", cmd.String())

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		log.Printf("❌ Errore durante l'esecuzione di meshtastic-go: %v\nOutput:\n%s", err, out.String())
		return nil, err
	}

	var parsed meshInfo
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		log.Printf("❌ Errore parsing JSON: %v\nOutput:\n%s", err, out.String())
		return nil, err
	}

	info := &Info{
		Firmware: parsed.DeviceMetrics.FirmwareVersion,
	}

	// trova il nome del nodo confrontando my_node_num con gli utenti
	if user, ok := parsed.User[parsed.MyNodeNum]; ok {
		info.NodeName = user.LongName
	} else {
		// fallback: prova a cercare anche con versione int del my_node_num
		if idInt, err := strconv.ParseInt(parsed.MyNodeNum, 0, 64); err == nil {
			hexID := "!" + strconv.FormatInt(idInt, 16)
			for k, u := range parsed.User {
				if k == hexID {
					info.NodeName = u.LongName
					break
				}
			}
		}
	}

	log.Printf("✅ Info trovate - Nodo: %s, Firmware: %s\n", info.NodeName, info.Firmware)
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
