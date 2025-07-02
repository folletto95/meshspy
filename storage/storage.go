package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        text TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Add(text string) error {
	_, err := s.db.Exec(`INSERT INTO messages(text) VALUES (?)`, text)
	return err
}

func (s *Store) List() ([]string, error) {
	rows, err := s.db.Query(`SELECT text FROM messages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		msgs = append(msgs, t)
	}
	return msgs, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}