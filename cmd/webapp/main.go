package main

import (
	"log"
	"net/http"
	"os"

	mqttpkg "meshspy/client"
	"meshspy/config"

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		token := client.Subscribe(cfg.MQTTTopic, 0, func(c mqtt.Client, m mqtt.Message) {
			conn.WriteMessage(websocket.TextMessage, m.Payload())
		})
		token.Wait()
		if token.Error() != nil {
			log.Printf("MQTT subscribe error: %v", token.Error())
			return
		}
		defer client.Unsubscribe(cfg.MQTTTopic)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("websocket read error: %v", err)
				break
			}
			t := client.Publish(cfg.CommandTopic, 0, false, message)
			t.Wait()
			if t.Error() != nil {
				log.Printf("MQTT publish error: %v", t.Error())
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
