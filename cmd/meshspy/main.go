package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
    "os/exec"

	"meshspy/config"
	"meshspy/mqtt"
	"meshspy/serial"
)

func main() {
	// Carica la configurazione dalle variabili d'ambiente
	cfg := config.Load()

	// Connessione al broker MQTT
	client, err := mqtt.ConnectMQTT(cfg)
	if err != nil {
		log.Fatalf("❌ Errore connessione MQTT: %v", err)
	}
	defer client.Disconnect(250)

	// Inizializza il canale di uscita per la gestione dei segnali di terminazione
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

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

	// 📡 Stampa info da meshtastic-go (se disponibile)
	info, err := mqtt.GetInfo(cfg.SerialPort)
	if err != nil {
		log.Printf("⚠️ Errore ottenimento info meshtastic-go: %v", err)
	} else {
		fmt.Printf("ℹ️  Info dispositivo Meshtastic:\n%s\n", info)
	}

	// Mantieni il programma in esecuzione finché non ricevi un segnale di uscita
	<-sigs
	log.Println("👋 Uscita in corso...")
	time.Sleep(time.Second)
}
