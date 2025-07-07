package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	mqttpkg "meshspy/client"

	_ "github.com/mattn/go-sqlite3"
)

// NodeStore manages persistent storage of NodeInfo records.
type NodeStore struct {
	db *sql.DB
}

// NodePosition represents a recorded position for a node.
type NodePosition struct {
	NodeID     string
	Latitude   float64
	Longitude  float64
	Altitude   int
	Time       int64
	ReceivedAt time.Time
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
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS positions (
        node_id TEXT,
        latitude REAL,
        longitude REAL,
        altitude INTEGER,
        time INTEGER,
        received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
	if err != nil {
		return err
	}
	return s.addPosition(info)
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

// addPosition stores the location from info if available.
func (s *NodeStore) addPosition(info *mqttpkg.NodeInfo) error {
	if info == nil {
		return nil
	}
	if info.Latitude == 0 && info.Longitude == 0 {
		return nil
	}
	_, err := s.db.Exec(`INSERT INTO positions(node_id, latitude, longitude, altitude, time, received_at)
                VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		info.ID, info.Latitude, info.Longitude, info.Altitude, info.LocationTime)
	return err
}

// Positions returns recorded positions. When nodeID is empty all positions are returned.
func (s *NodeStore) Positions(nodeID string) ([]NodePosition, error) {
	var rows *sql.Rows
	var err error
	if nodeID == "" {
		rows, err = s.db.Query(`SELECT node_id, latitude, longitude, altitude, time, received_at FROM positions ORDER BY received_at`)
	} else {
		rows, err = s.db.Query(`SELECT node_id, latitude, longitude, altitude, time, received_at FROM positions WHERE node_id = ? ORDER BY received_at`, nodeID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []NodePosition
	for rows.Next() {
		var p NodePosition
		if err := rows.Scan(&p.NodeID, &p.Latitude, &p.Longitude, &p.Altitude, &p.Time, &p.ReceivedAt); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}
