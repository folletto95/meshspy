package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"
)

type server struct {
	mu       sync.Mutex
	messages []string
	tmpl     *template.Template
}

func newServer() *server {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	return &server{tmpl: tmpl}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err == nil {
			msg := r.Form.Get("msg")
			if msg != "" {
				s.mu.Lock()
				s.messages = append(s.messages, msg)
				s.mu.Unlock()
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.mu.Lock()
	data := struct{ Messages []string }{append([]string(nil), s.messages...)}
	s.mu.Unlock()
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
	srv := newServer()
	http.HandleFunc("/", srv.handleIndex)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}