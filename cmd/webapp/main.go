package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	mqttpkg "meshspy/client"
	"meshspy/config"
	"meshspy/storage"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{}

func main() {
	// Load environment variables from .env.runtime if present
	if err := godotenv.Load(".env.runtime"); err != nil {
		log.Printf("⚠️  Nessun file .env.runtime trovato o errore di caricamento: %v", err)
	}

	cfg := config.Load()

	client, err := mqttpkg.ConnectMQTT(cfg)
	if err != nil {
		log.Fatalf("MQTT connect error: %v", err)
	}
	defer client.Disconnect(250)

	dbPath := os.Getenv("NODE_DB_PATH")
	if dbPath == "" {
		dbPath = "nodes.db"
	}
	nodeStore, err := storage.NewNodeStore(dbPath)
	if err != nil {
		log.Fatalf("node store open error: %v", err)
	}
	defer nodeStore.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("web/index.html"); err == nil {
			http.ServeFile(w, r, "web/index.html")
		} else {
			http.ServeFile(w, r, "cmd/webapp/index.html")
		}
	})

	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		nodes, err := nodeStore.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}

		// Channel used to queue messages coming from the websocket before
		// publishing them to MQTT. The small buffer (10) prevents slow
		// MQTT publishes from blocking the websocket reader.
		sendCh := make(chan []byte, 10)

		log.Printf("🔌 websocket client %s connected", r.RemoteAddr)
		defer func() {
			log.Printf("🔌 websocket client %s disconnected", r.RemoteAddr)
			close(sendCh) // stop publisher goroutine
			conn.Close()
		}()

		// Goroutine responsible for publishing messages received from the
		// websocket to MQTT. It exits when sendCh is closed.
		go func() {
			for msg := range sendCh {
				t := client.Publish(cfg.CommandTopic, 0, false, msg)
				// Wait for at most 3 seconds for publish to complete.
				if !t.WaitTimeout(3 * time.Second) {
					log.Printf("MQTT publish timeout")
					continue
				}
				if t.Error() != nil {
					log.Printf("MQTT publish error: %v", t.Error())
				} else {
					log.Printf("⬆️ MQTT publish to %s: %s", cfg.CommandTopic, msg)
				}
			}
		}()

		token := client.Subscribe(cfg.MQTTTopic, 0, func(c mqtt.Client, m mqtt.Message) {
			log.Printf("⬇️ MQTT message on %s: %s", cfg.MQTTTopic, m.Payload())
			if err := conn.WriteMessage(websocket.TextMessage, m.Payload()); err != nil {
				log.Printf("websocket write error: %v", err)
			}
		})
		token.Wait()
		if token.Error() != nil {
			log.Printf("MQTT subscribe error: %v", token.Error())
			return
		}
		log.Printf("✅ subscribed to %s", cfg.MQTTTopic)
		defer func() {
			client.Unsubscribe(cfg.MQTTTopic)
			log.Printf("🔕 unsubscribed from %s", cfg.MQTTTopic)
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("websocket read error: %v", err)
				break
			}
			log.Printf("➡️  from web client: %s", message)
			// Queue the message for publishing. If the buffer is full
			// the message is dropped to avoid blocking the websocket
			// reader.
			select {
			case sendCh <- message:
			default:
				log.Printf("publish queue full, dropping message")
			}
			if err := conn.WriteMessage(websocket.TextMessage, append([]byte("echo: "), message...)); err != nil {
				log.Printf("websocket echo error: %v", err)
			}
		}
	})

	port := os.Getenv("WEB_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🌐 Web app in ascolto su :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
