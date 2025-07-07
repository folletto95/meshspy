package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	mqttpkg "meshspy/client"
	"meshspy/config"
	"meshspy/storage"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

// apiServer holds dependencies for the HTTP API.
type apiServer struct {
	mqtt  mqtt.Client
	cfg   config.Config
	store *storage.NodeStore
}

func newServer(m mqtt.Client, cfg config.Config, store *storage.NodeStore) *apiServer {
	return &apiServer{mqtt: m, cfg: cfg, store: store}
}

// listNodes returns the known nodes as JSON.
func (s *apiServer) listNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// sendCommand publishes a command payload on the MQTT command topic.
func (s *apiServer) sendCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Cmd string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	token := s.mqtt.Publish(s.cfg.CommandTopic, 0, false, req.Cmd)
	token.Wait()
	if token.Error() != nil {
		http.Error(w, token.Error().Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	if err := godotenv.Load(".env.runtime"); err != nil {
		log.Printf("⚠️  .env.runtime not loaded: %v", err)
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
	store, err := storage.NewNodeStore(dbPath)
	if err != nil {
		log.Fatalf("node store open error: %v", err)
	}
	defer store.Close()

	srv := newServer(client, cfg, store)
	http.HandleFunc("/api/nodes", srv.listNodes)
	http.HandleFunc("/api/send", srv.sendCommand)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("🌐 Management server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
