package main

import (
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
	"meshspy/serial"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const welcomeMessage = "Ciao da MeshSpy, presto (spero) per tutti"

func main() {
	log.Println("🔥 MeshSpy avviamento iniziato...")

	// Carica .env.runtime se presente
	if err := godotenv.Load(".env.runtime"); err != nil {
		log.Printf("⚠️  Nessun file .env.runtime trovato o errore di caricamento: %v", err)
	}

	log.Println("🚀 MeshSpy avviato con successo! Inizializzazione in corso..")

	// Carica la configurazione dalle variabili d'ambiente
	cfg := config.Load()

	// Connessione al broker MQTT
	client, err := mqttpkg.ConnectMQTT(cfg)
	if err != nil {
		log.Fatalf("❌ Errore connessione MQTT: %v", err)
	}
	defer client.Disconnect(250)

	if err := mqttpkg.PublishAlive(client, cfg.MQTTTopic); err != nil {
		log.Printf("⚠️  Errore invio messaggio Alive: %v", err)
	} else {
		log.Printf("✅ Messaggio di test inviato su '%s'", cfg.MQTTTopic)
	}

	// Sottoscrivi al topic dei comandi e inoltra i messaggi sulla seriale
	token := client.Subscribe(cfg.CommandTopic, 0, func(c paho.Client, m paho.Message) {
		msg := string(m.Payload())
		switch {
		case msg == "sendhello":
			if err := serial.SendTextMessage(cfg.SerialPort, welcomeMessage); err != nil {
				log.Printf("❌ Errore invio messaggio standard: %v", err)
			} else {
				log.Printf("✅ Messaggio standard inviato")
			}
		case strings.HasPrefix(msg, "send:"):
			text := strings.TrimPrefix(msg, "send:")
			if err := serial.SendTextMessage(cfg.SerialPort, text); err != nil {
				log.Printf("❌ Errore invio messaggio personalizzato: %v", err)
			} else {
				log.Printf("✅ Messaggio personalizzato inviato: %s", text)
			}
		default:
			if err := serial.Send(cfg.SerialPort, cfg.BaudRate, msg); err != nil {
				log.Printf("❌ Errore invio seriale: %v", err)
			} else {
				log.Printf("➡️  Comando inoltrato alla seriale: %s", m.Payload())
			}
		}
	})
	token.Wait()
	if token.Error() != nil {
		log.Printf("⚠️  Errore sottoscrizione comandi: %v", token.Error())
	}

	// Inizializza il canale di uscita per la gestione dei segnali di terminazione
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 📡 Attende la disponibilità della porta seriale prima di eseguire meshtastic-go
	if err := serial.WaitForSerial(cfg.SerialPort, 30*time.Second); err != nil {
		log.Fatalf("❌ Porta seriale %s non disponibile: %v", cfg.SerialPort, err)
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

	if info, err := mqttpkg.GetLocalNodeInfo(cfg.SerialPort); err == nil {
		if err := mqttpkg.SaveNodeInfo(info, "nodes.json"); err != nil {
			log.Printf("⚠️ Salvataggio info nodo fallito: %v", err)
		}
		cfgFile := mqttpkg.BuildConfigFilename(info)
		if err := mqttpkg.ExportConfig(cfg.SerialPort, cfgFile); err != nil {
			log.Printf("⚠️ Esportazione configurazione fallita: %v", err)
		} else {
			log.Printf("✅ Configurazione salvata in %s", cfgFile)
		}
		if err := serial.SendTextMessage(cfg.SerialPort, welcomeMessage); err != nil {
			log.Printf("⚠️ Errore invio messaggio di benvenuto: %v", err)
		} else {
			log.Printf("✅ Messaggio di benvenuto inviato")
		}
	} else {
		log.Printf("⚠️ Lettura info nodo fallita: %v", err)
	}

	// Avvia la lettura dalla porta seriale in un goroutine
	go func() {
		serial.ReadLoop(cfg.SerialPort, cfg.BaudRate, cfg.Debug, func(data string) {
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
