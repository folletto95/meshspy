package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv" // ← used to read .env files

	mqttpkg "meshspy/client"
	"meshspy/config"
	"meshspy/mgmtapi"
	"meshspy/nodemap"
	latestpb "meshspy/proto/latest/meshtastic"
	"meshspy/serial"
	"meshspy/storage"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	welcomeMessage = "Ciao da MeshSpy, presto (spero) per tutti"
	aliveMessage   = "MeshSpy Alive"
	// logFilename specifies the file where all log output is written
	logFilename = "log.txt"
)

// Version of the MeshSpy program. This value can be overridden at build time
// using: go build -ldflags="-X main.Version=x.y.z"
var Version = "dev"

func init() {
	// Load the version from the .env.build file if present
	if err := godotenv.Load(".env.build"); err == nil {
		if Version == "dev" {
			if v := os.Getenv("MESHSPY_VERSION"); v != "" {
				Version = v
			}
		}
	}
}

func main() {
	// Open or create the log file and direct all log output to it
	f, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("unable to open log file: %v", err)
	}
	defer f.Close()
	// Write logs to both stdout and the file for easier debugging
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	log.Println("🔥 MeshSpy avviamento iniziato...")
	log.Printf("📦 Versione MeshSpy: %s", Version)

	msg := flag.String("sendtext", "", "Messaggio da inviare invece di avviare il listener")
	dest := flag.String("dest", "", "Nodo destinatario (opzionale)")
	flag.Parse()

	// Load .env.runtime if present
	if err := godotenv.Load(".env.runtime"); err != nil {
		log.Printf("⚠️  Nessun file .env.runtime trovato o errore di caricamento: %v", err)
	}

	log.Println("🚀 MeshSpy avviato con successo! Inizializzazione in corso..")

	// Load configuration from environment variables
	cfg := config.Load()
	nodes := nodemap.New()
	mgmt := mgmtapi.New(cfg.MgmtURL)

	// Print MQTT credentials so they can be verified before connecting
	log.Printf("ℹ️  MQTT user: %s", cfg.User)
	log.Printf("ℹ️  MQTT password: %s", cfg.Password)

	nodeDB := os.Getenv("NODE_DB_PATH")
	if nodeDB == "" {
		nodeDB = "nodes.db"
	}
	nodeStore, err := storage.NewNodeStore(nodeDB)
	if err != nil {
		log.Fatalf("❌ apertura db nodi: %v", err)
	}
	defer nodeStore.Close()

	if *msg != "" {
		if err := serial.SendText(cfg.SerialPort, *msg); err != nil {
			log.Fatalf("❌ Errore invio messaggio: %v", err)
		}
		if *dest != "" {
			log.Printf("✅ Messaggio inviato a %s", *dest)
		} else {
			log.Printf("✅ Messaggio inviato")
		}
		return
	}

	// Connect to the MQTT broker
	client, err := mqttpkg.ConnectMQTT(cfg)
	if err != nil {
		log.Fatalf("❌ Errore connessione MQTT: %v", err)
	}
	defer client.Disconnect(250)

	if err := mqttpkg.SendAliveIfNeeded(client, cfg); err != nil {
		log.Printf("⚠️  Errore invio messaggio Alive: %v", err)
	} else if cfg.SendAlive {
		log.Printf("✅ Messaggio Alive inviato su '%s'", cfg.MQTTTopic)
	}

	// Subscribe to the command topic and forward messages over serial
	var portMgr *serial.Manager

	token := client.Subscribe(cfg.CommandTopic, 0, func(c paho.Client, m paho.Message) {
		msg := string(m.Payload())
		log.Printf("📥 comando ricevuto (%s): %s", m.Topic(), msg)
		if portMgr == nil {
			log.Printf("❌ Porta seriale non inizializzata")
			return
		}
		switch {
		case msg == "sendhello":
			if err := portMgr.SendTextMessage(welcomeMessage); err != nil {
				log.Printf("❌ Errore invio messaggio standard: %v", err)
			} else {
				log.Printf("✅ Messaggio standard inviato")
			}
		case strings.HasPrefix(msg, "send:"):
			text := strings.TrimPrefix(msg, "send:")
			if err := portMgr.SendTextMessage(text); err != nil {
				log.Printf("❌ Errore invio messaggio personalizzato: %v", err)
			} else {
				log.Printf("✅ Messaggio personalizzato inviato: %s", text)
			}
		default:
			if err := portMgr.SendTextMessage(msg); err != nil {
				log.Printf("❌ Errore invio messaggio: %v", err)
			} else {
				log.Printf("✅ Messaggio inviato: %s", msg)
			}
		}
	})
	token.Wait()
	if token.Error() != nil {
		log.Printf("⚠️  Errore sottoscrizione comandi: %v", token.Error())
	}
	log.Printf("✅ in ascolto su topic comandi %s", cfg.CommandTopic)

	// Initialize the exit channel to handle termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 📡 Wait for the serial port to be available before running meshtastic-go
	if err := serial.WaitForSerial(cfg.SerialPort, 30*time.Second); err != nil {
		log.Fatalf("❌ Porta seriale %s non disponibile: %v", cfg.SerialPort, err)
	}

	// Send an Alive message to the node if requested
	if cfg.SendAlive {
		if err := serial.SendTextMessage(cfg.SerialPort, aliveMessage); err != nil {
			log.Printf("⚠️  Errore invio messaggio Alive al nodo: %v", err)
		} else {
			log.Printf("✅ Messaggio Alive inviato al nodo")
		}
	}

	// 📡 Print info from meshtastic-go (if available)

	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", cfg.SerialPort, "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("⚠️ Errore ottenimento info meshtastic-go: %v", err)
		if len(output) > 0 {
			log.Printf("Output meshtastic-go:\n%s", string(output))
		}
	} else {
		fmt.Printf("ℹ️  Info dispositivo Meshtastic:\n%s\n", output)
	}

	var protoVer string
	info, err := mqttpkg.GetLocalNodeInfoCached(cfg.SerialPort, "nodes.json")
	if err != nil {
		log.Printf("⚠️ Lettura info nodo fallita: %v", err)
	} else {
		protoVer = mqttpkg.ProtoVersionForFirmware(info.FirmwareVersion)
		if err := mqttpkg.SaveNodeInfo(info, "nodes.json"); err != nil {
			log.Printf("⚠️ Salvataggio info nodo fallito: %v", err)
		}
		nodes.Update(info.Num, info.LongName, info.ShortName)
		if err := nodeStore.Upsert(info); err != nil {
			log.Printf("⚠️ aggiornamento db nodi: %v", err)
		}
		if err := mgmt.SendNode(info); err != nil {
			log.Printf("⚠️ invio info nodo al server: %v", err)
		}
		if nodesList, err := mqttpkg.GetMeshNodes(cfg.SerialPort); err == nil {
			for _, n := range nodesList {
				nodes.Update(n.Num, n.LongName, n.ShortName)
				if err := nodeStore.Upsert(n); err != nil {
					log.Printf("⚠️ aggiornamento db nodi: %v", err)
				}
				if err := mgmt.SendNode(n); err != nil {
					log.Printf("⚠️ invio info nodo al server: %v", err)
				}
			}
		} else {
			log.Printf("⚠️ lettura nodi fallita: %v", err)
		}
		cfgFile := mqttpkg.BuildConfigFilename(info)
		if err := mqttpkg.ExportConfig(cfg.SerialPort, cfgFile); err != nil {
			log.Printf("⚠️ Esportazione configurazione fallita: %v", err)
		} else {
			log.Printf("✅ Configurazione salvata in %s", cfgFile)
		}
	}
	if err := serial.SendTextMessage(cfg.SerialPort, welcomeMessage); err != nil {
		log.Printf("⚠️ Errore invio messaggio di benvenuto: %v", err)
	} else {
		log.Printf("✅ Messaggio di benvenuto inviato")
	}

	portMgr, err = serial.OpenManager(cfg.SerialPort, cfg.BaudRate, protoVer)
	if err != nil {
		log.Fatalf("❌ apertura porta seriale: %v", err)
	}
	defer portMgr.Close()

	// Start reading from the serial port in a goroutine
	go func() {
		portMgr.ReadLoop(cfg.Debug, protoVer, nodes, func(ni *latestpb.NodeInfo) {
			info := mqttpkg.NodeInfoFromProto(ni)
			if info != nil {
				if err := nodeStore.Upsert(info); err != nil {
					log.Printf("⚠️ aggiornamento db nodi: %v", err)
				}
				if err := mgmt.SendNode(info); err != nil {
					log.Printf("⚠️ invio info nodo al server: %v", err)
				}
			}
		}, func(mi *latestpb.MyNodeInfo) {
			info := mqttpkg.NodeInfoFromMyInfo(mi)
			if info != nil {
				if err := nodeStore.Upsert(info); err != nil {
					log.Printf("⚠️ aggiornamento db nodi: %v", err)
				}
				if err := mgmt.SendNode(info); err != nil {
					log.Printf("⚠️ invio info nodo al server: %v", err)
				}
			}
		}, func(tm *latestpb.Telemetry) {
			b, _ := json.Marshal(tm)
			log.Printf("📊 Telemetry: %s", string(b))
		}, func(wp *latestpb.Waypoint) {
			if err := nodeStore.AddWaypoint(wp); err != nil {
				log.Printf("⚠️ salvataggio waypoint: %v", err)
			}
		}, func(adm []byte) {
			log.Printf("⚙️ Admin: %x", adm)
		}, func(alert string) {
			log.Printf("🚨 Alert: %s", alert)
		}, func(txt string) {
			log.Printf("💬 Text: %s", txt)
		}, func(data string) {

			// Publish every received message on the MQTT topic
			token := client.Publish(cfg.MQTTTopic, 0, false, data)
			token.Wait()
			if token.Error() != nil {
				log.Printf("❌ Errore pubblicazione MQTT: %v", token.Error())
			} else {
				log.Printf("📡 Dato pubblicato su '%s': %s", cfg.MQTTTopic, data)
			}
		})
	}()

	// Keep the program running until an exit signal is received
	<-sigs
	log.Println("👋 Uscita in corso...")
	time.Sleep(time.Second)
}
