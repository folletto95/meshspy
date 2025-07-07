// Package mqtt provides utilities to parse Meshtastic node information and publish it to an MQTT broker.
package mqtt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"meshspy/config"
)

// NodeInfo represents the information extracted from a Meshtastic device.
type NodeInfo struct {
	ID                    string
	Num                   uint32
	LongName              string
	ShortName             string
	MacAddr               string
	HwModel               string
	Role                  string
	Latitude              float64
	Longitude             float64
	Altitude              int
	LocationTime          int64
	LocationSource        string
	BatteryLevel          int
	Voltage               float64
	ChannelUtil           float64
	AirUtilTx             float64
	UptimeSeconds         int
	FirmwareVersion       string
	DeviceStateVer        int
	CanShutdown           bool
	HasWifi               bool
	HasBluetooth          bool
	HasEthernet           bool
	RadioRole             string
	PositionFlags         int
	RadioHwModel          string
	HasRemoteHardware     bool
	Snr                   float64
	LastHeard             int64
	Channel               int
	ViaMqtt               bool
	HopsAway              int
	IsFavorite            bool
	IsIgnored             bool
	IsKeyManuallyVerified bool
}

// Espressioni regolari
//var (
//	nameRe = regexp.MustCompile(`long_name:"([^"]+)"`)
//	fwRe   = regexp.MustCompile(`FirmwareVersion\s+([^\s]+)`)
//)

// GetLocalNodeInfo esegue meshtastic-go e recupera i dati dal primo nodo dopo "Radio Settings:"
func GetLocalNodeInfo(port string) (*NodeInfo, error) {
	var (
		output []byte
		err    error
	)

	for attempt := 1; attempt <= 5; attempt++ {
		cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "info")
		output, err = cmd.CombinedOutput()
		//fmt.Printf("📤 Eseguo comando: %s\n", strings.Join(cmd.Args, " "))
		//fmt.Println("🔍 Output completo di meshtastic-go:\n")
		//fmt.Println(string(output))

		if err == nil {
			break
		}

		fmt.Printf("❌ Errore durante l'esecuzione di meshtastic-go (tentativo %d/5): %v\n", attempt, err)
		time.Sleep(time.Second)
	}

	if err != nil {
		fmt.Printf("❌ Errore durante l'esecuzione di meshtastic-go: %v\n", err)
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var inNodeInfo bool
	node := &NodeInfo{}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Node Info") {
			if !inNodeInfo {
				inNodeInfo = true
				continue
			} else {
				break // fermiamoci al secondo nodo
			}
		}

		if !inNodeInfo {
			continue
		}

		switch {
		case strings.HasPrefix(line, "User"):
			re := regexp.MustCompile(`id:"([^"]+)".*?long_name:"([^"]+)".*?short_name:"([^"]+)".*?macaddr:"([^"]+)".*?hw_model:(\S+).*?role:(\S+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 7 {
				node.ID = matches[1]
				node.LongName = matches[2]
				node.ShortName = matches[3]
				node.MacAddr = matches[4]
				node.HwModel = matches[5]
				node.Role = matches[6]
			}
		case strings.HasPrefix(line, "Position"):
			re := regexp.MustCompile(`latitude_i:(\d+)\s+longitude_i:(\d+)\s+altitude:(-?\d+)\s+time:(\d+)\s+location_source:(\S+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 6 {
				lat, _ := strconv.Atoi(matches[1])
				lon, _ := strconv.Atoi(matches[2])
				node.Latitude = float64(lat) / 1e7
				node.Longitude = float64(lon) / 1e7
				node.Altitude, _ = strconv.Atoi(matches[3])
				node.LocationTime, _ = strconv.ParseInt(matches[4], 10, 64)
				node.LocationSource = matches[5]
			}
		case strings.HasPrefix(line, "DeviceMetrics"):
			re := regexp.MustCompile(`battery_level:(\d+)\s+voltage:(\S+)\s+channel_utilization:(\S+)\s+air_util_tx:(\S+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 5 {
				node.BatteryLevel, _ = strconv.Atoi(matches[1])
				node.Voltage, _ = strconv.ParseFloat(matches[2], 64)
				node.ChannelUtil, _ = strconv.ParseFloat(matches[3], 64)
				node.AirUtilTx, _ = strconv.ParseFloat(matches[4], 64)
			}
		case strings.HasPrefix(line, "FirmwareVersion"):
			node.FirmwareVersion = strings.TrimSpace(strings.TrimPrefix(line, "FirmwareVersion"))
		case strings.HasPrefix(line, "DeviceStateVersion"):
			node.DeviceStateVer, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "DeviceStateVersion")))
		case strings.HasPrefix(line, "CanShutdown"):
			node.CanShutdown = strings.Contains(line, "true")
		case strings.HasPrefix(line, "HasWifi"):
			node.HasWifi = strings.Contains(line, "true")
		case strings.HasPrefix(line, "HasBluetooth"):
			node.HasBluetooth = strings.Contains(line, "true")
		case strings.HasPrefix(line, "HasEthernet"):
			node.HasEthernet = strings.Contains(line, "true")
		case strings.HasPrefix(line, "Role"):
			node.RadioRole = strings.TrimSpace(strings.TrimPrefix(line, "Role"))
		case strings.HasPrefix(line, "PositionFlags"):
			node.PositionFlags, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "PositionFlags")))
		case strings.HasPrefix(line, "HwModel"):
			node.RadioHwModel = strings.TrimSpace(strings.TrimPrefix(line, "HwModel"))
		case strings.HasPrefix(line, "HasRemoteHardware"):
			node.HasRemoteHardware = strings.Contains(line, "true")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return node, nil
}

// ConnectMQTT crea e restituisce un client MQTT connesso
func ConnectMQTT(cfg config.Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	if cfg.User != "" {
		opts.SetUsername(cfg.User)
		opts.SetPassword(cfg.Password)
	}

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("⚠️ MQTT connection lost: %v", err)
	})
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("✅ MQTT connection established")
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	return client, token.Error()
}

// SaveNodeInfo serializes NodeInfo to a JSON file
func SaveNodeInfo(info *NodeInfo, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(info)
}

// LoadNodeInfo deserializes NodeInfo from a JSON file
func LoadNodeInfo(path string) (*NodeInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var info NodeInfo
	if err := json.NewDecoder(f).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetLocalNodeInfoCached loads node info from the given path when available.
// When the file does not exist or cannot be parsed, it executes
// meshtastic-go to retrieve the information and saves it for later use.
func GetLocalNodeInfoCached(port, path string) (*NodeInfo, error) {
	if info, err := LoadNodeInfo(path); err == nil {
		if info.LongName != "" && info.FirmwareVersion != "" {
			return info, nil
		}
	}

	info, err := GetLocalNodeInfo(port)
	if err != nil {
		return nil, err
	}
	if err := SaveNodeInfo(info, path); err != nil {
		fmt.Printf("⚠️ salvataggio info nodo fallito: %v\n", err)
	}
	return info, nil
}

// PublishAlive sends a simple \"MeshSpy Alive\" message to the given topic.
func PublishAlive(client mqtt.Client, topic string) error {
	token := client.Publish(topic, 0, false, []byte("MeshSpy Alive"))
	token.Wait()
	return token.Error()
}

// SendAliveIfNeeded publishes an Alive message when cfg.SendAlive is true.
func SendAliveIfNeeded(client mqtt.Client, cfg config.Config) error {
	if !cfg.SendAlive {
		return nil
	}
	return PublishAlive(client, cfg.MQTTTopic)
}
