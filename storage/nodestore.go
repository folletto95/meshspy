package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"time"

	mqttpkg "meshspy/client"
	latestpb "meshspy/proto/latest/meshtastic"

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

// TelemetryRecord represents stored device metrics with timestamps.
type TelemetryRecord struct {
	BatteryLevel       uint32
	Voltage            float64
	ChannelUtilization float64
	AirUtilTx          float64
	UptimeSeconds      uint32
	Time               uint32
	ReceivedAt         time.Time
}

// NewNodeStore opens or creates a SQLite database at path and prepares the nodes table.
func NewNodeStore(path string) (*NodeStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
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
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS telemetry (
        battery_level INTEGER,
        voltage REAL,
        channel_utilization REAL,
        air_util_tx REAL,
        uptime_seconds INTEGER,
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

// AddWaypoint stores the coordinates from a Waypoint message in the positions table.
func (s *NodeStore) AddWaypoint(wp *latestpb.Waypoint) error {
	if wp == nil || wp.LatitudeI == nil || wp.LongitudeI == nil {
		return nil
	}
	id := fmt.Sprintf("0x%x", wp.GetId())
	lat := float64(wp.GetLatitudeI()) / 1e7
	lon := float64(wp.GetLongitudeI()) / 1e7
	_, err := s.db.Exec(`INSERT INTO positions(node_id, latitude, longitude, altitude, time, received_at)
                VALUES(?, ?, ?, 0, 0, CURRENT_TIMESTAMP)`, id, lat, lon)
	return err
}

// AddTelemetry stores telemetry metrics in the database. Only DeviceMetrics
// from the Telemetry message are saved when present.
func (s *NodeStore) AddTelemetry(tel *latestpb.Telemetry) error {
	if tel == nil {
		return nil
	}
	dm := tel.GetDeviceMetrics()
	if dm == nil {
		return nil
	}
	_, err := s.db.Exec(`INSERT INTO telemetry(
                battery_level, voltage, channel_utilization, air_util_tx,
                uptime_seconds, time, received_at)
                VALUES(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		dm.GetBatteryLevel(), dm.GetVoltage(), dm.GetChannelUtilization(),
		dm.GetAirUtilTx(), dm.GetUptimeSeconds(), tel.GetTime())
	return err
}

// Telemetry returns stored telemetry records ordered by the time they were
// received.
func (s *NodeStore) Telemetry() ([]TelemetryRecord, error) {
	rows, err := s.db.Query(`SELECT battery_level, voltage, channel_utilization,
                air_util_tx, uptime_seconds, time, received_at FROM telemetry
                ORDER BY received_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []TelemetryRecord
	for rows.Next() {
		var r TelemetryRecord
		if err := rows.Scan(&r.BatteryLevel, &r.Voltage, &r.ChannelUtilization,
			&r.AirUtilTx, &r.UptimeSeconds, &r.Time, &r.ReceivedAt); err != nil {
			return nil, err
		}
		recs = append(recs, r)
	}
	return recs, rows.Err()
}
