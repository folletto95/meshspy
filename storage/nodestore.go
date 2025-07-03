package storage

import (
	"database/sql"
	"encoding/json"

	mqttpkg "meshspy/client"

	_ "github.com/mattn/go-sqlite3"
)

// NodeStore manages persistent storage of NodeInfo records.
type NodeStore struct {
	db *sql.DB
}

// NewNodeStore opens or creates a SQLite database at path and prepares the nodes table.
func NewNodeStore(path string) (*NodeStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS nodes (
        id TEXT PRIMARY KEY,
        info TEXT,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`); err != nil {
		db.Close()
		return nil, err
	}
	return &NodeStore{db: db}, nil
}

// Close closes the underlying database.
func (s *NodeStore) Close() error {
	return s.db.Close()
}

// Upsert inserts or updates the given NodeInfo in the database.
func (s *NodeStore) Upsert(info *mqttpkg.NodeInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO nodes(id, info, updated_at) VALUES(?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(id) DO UPDATE SET info=excluded.info, updated_at=CURRENT_TIMESTAMP`,
		info.ID, string(data))
	return err
}

// List returns all NodeInfo records stored in the database.
func (s *NodeStore) List() ([]*mqttpkg.NodeInfo, error) {
	rows, err := s.db.Query(`SELECT info FROM nodes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*mqttpkg.NodeInfo
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var n mqttpkg.NodeInfo
		if err := json.Unmarshal([]byte(raw), &n); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}
