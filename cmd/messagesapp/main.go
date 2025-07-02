package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"meshspy/storage"
)

type server struct {
	mu    sync.Mutex
	store *storage.Store
	tmpl  *template.Template
}

func newServer(store *storage.Store) *server {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	return &server{tmpl: tmpl, store: store}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err == nil {
			msg := r.Form.Get("msg")
			if msg != "" {
				if err := s.store.Add(msg); err != nil {
					log.Printf("store add: %v", err)
				}
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	msgs, err := s.store.List()
	if err != nil {
		log.Printf("store list: %v", err)
	}
	data := struct{ Messages []string }{Messages: msgs}
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("template execute: %v", err)
	}
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Message Board</title>
</head>
<body>
<h1>Message Board</h1>
<form method="POST" action="/">
<input type="text" name="msg" placeholder="Type a message">
<button type="submit">Send</button>
</form>
<ul>
{{range .Messages}}<li>{{.}}</li>{{end}}
</ul>
</body>
</html>`

func main() {
	dbPath := os.Getenv("MSG_DB_PATH")
	if dbPath == "" {
		dbPath = "messages.db"
	}
	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	srv := newServer(store)
	http.HandleFunc("/", srv.handleIndex)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}