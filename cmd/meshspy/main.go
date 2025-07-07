package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv" // ← aggiunto per leggere il file .env

	mqttpkg "meshspy/client"
	"meshspy/config"
	"meshspy/gui"
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
)

// Version of the MeshSpy program. This value can be overridden at build time
// using: go build -ldflags="-X main.Version=x.y.z"
var Version = "dev"

func main() {
	log.Println("🔥 MeshSpy avviamento iniziato...")
	log.Printf("📦 Versione MeshSpy: %s", Version)

	msg := flag.String("sendtext", "", "Messaggio da inviare invece di avviare il listener")
	dest := flag.String("dest", "", "Nodo destinatario (opzionale)")
	flag.Parse()

	// Carica .env.runtime se presente
	if err := godotenv.Load(".env.runtime"); err != nil {
		log.Printf("⚠️  Nessun file .env.runtime trovato o errore di caricamento: %v", err)
	}

	log.Println("🚀 MeshSpy avviato con successo! Inizializzazione in corso..")

	// Carica la configurazione dalle variabili d'ambiente
	cfg := config.Load()
	nodes := nodemap.New()
	mgmt := mgmtapi.New(cfg.MgmtURL)

	nodeDB := os.Getenv("NODE_DB_PATH")
	if nodeDB == "" {
		nodeDB = "nodes.db"
	}
	nodeStore, err := storage.NewNodeStore(nodeDB)
	if err != nil {
		log.Fatalf("❌ apertura db nodi: %v", err)
	}
	defer nodeStore.Close()

	if cfg.EnableGUI {
		go gui.Run(cfg, nodeStore)
	}

	if *msg != "" {
		if err := serial.SendTextMessageTo(cfg.SerialPort, *dest, *msg); err != nil {
			log.Fatalf("❌ Errore invio messaggio: %v", err)
		}
		log.Printf("✅ Messaggio inviato a %s", *dest)
		return
	}

	// Connessione al broker MQTT
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

	// Sottoscrivi al topic dei comandi e inoltra i messaggi sulla seriale
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

	// Inizializza il canale di uscita per la gestione dei segnali di terminazione
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 📡 Attende la disponibilità della porta seriale prima di eseguire meshtastic-go
	if err := serial.WaitForSerial(cfg.SerialPort, 30*time.Second); err != nil {
		log.Fatalf("❌ Porta seriale %s non disponibile: %v", cfg.SerialPort, err)
	}

	// Invia un messaggio Alive anche al nodo se richiesto
	if cfg.SendAlive {
		if err := serial.SendTextMessage(cfg.SerialPort, aliveMessage); err != nil {
			log.Printf("⚠️  Errore invio messaggio Alive al nodo: %v", err)
		} else {
			log.Printf("✅ Messaggio Alive inviato al nodo")
		}
	}

	// 📡 Stampa info da meshtastic-go (se disponibile)

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

	portMgr, err = serial.OpenManager(cfg.SerialPort, cfg.BaudRate)
	if err != nil {
		log.Fatalf("❌ apertura porta seriale: %v", err)
	}
	defer portMgr.Close()

	// Avvia la lettura dalla porta seriale in un goroutine
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
		}, nil, nil, func(data string) {
			// Pubblica ogni messaggio ricevuto sul topic MQTT
			token := client.Publish(cfg.MQTTTopic, 0, false, data)
			token.Wait()
			if token.Error() != nil {
				log.Printf("❌ Errore pubblicazione MQTT: %v", token.Error())
			} else {
				log.Printf("📡 Dato pubblicato su '%s': %s", cfg.MQTTTopic, data)
			}
		})
	}()

	// Mantieni il programma in esecuzione finché non ricevi un segnale di uscita
	<-sigs
	log.Println("👋 Uscita in corso...")
	time.Sleep(time.Second)
}
