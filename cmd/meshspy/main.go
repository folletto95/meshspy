package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv" // ← aggiunto per leggere il file .env

	"meshspy/client"
	"meshspy/config"
	"meshspy/serial"
)

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
	client, err := mqtt.ConnectMQTT(cfg)
	if err != nil {
		log.Fatalf("❌ Errore connessione MQTT: %v", err)
	}
	defer client.Disconnect(250)

	if err := mqtt.PublishAlive(client, cfg.MQTTTopic); err != nil {
		log.Printf("⚠️  Errore invio messaggio Alive: %v", err)
	} else {
		log.Printf("✅ Messaggio di test inviato su '%s'", cfg.MQTTTopic)
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

	if info, err := mqtt.GetLocalNodeInfo(cfg.SerialPort); err == nil {
		if err := mqtt.SaveNodeInfo(info, "nodes.json"); err != nil {
			log.Printf("⚠️ Salvataggio info nodo fallito: %v", err)
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
