package storage

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"
	mqttpkg "meshspy/client"
	latestpb "meshspy/proto/latest/meshtastic"
)

func TestNewNodeStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nodes.db")

	ns, err := NewNodeStore(dbPath)
	if err != nil {
		t.Fatalf("NewNodeStore returned error: %v", err)
	}
	if err := ns.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file not created: %v", err)
	}
}

func TestNodeStoreUpsertAndList(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nodes.db")

	ns, err := NewNodeStore(dbPath)
	if err != nil {
		t.Fatalf("NewNodeStore returned error: %v", err)
	}
	defer func() {
		ns.Close()
	}()

	n1 := &mqttpkg.NodeInfo{ID: "abc", LongName: "Alice"}
	if err := ns.Upsert(n1); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	// Update existing record
	n2 := &mqttpkg.NodeInfo{ID: "abc", LongName: "Alice Updated"}
	if err := ns.Upsert(n2); err != nil {
		t.Fatalf("Upsert update returned error: %v", err)
	}

	nodes, err := ns.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	got := nodes[0]
	if got.LongName != n2.LongName || got.ID != n2.ID {
		t.Fatalf("node data mismatch: got %+v want %+v", got, n2)
	}
}

func TestNodeStoreUpsertInsertAndUpdate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nodes.db")

	ns, err := NewNodeStore(dbPath)
	if err != nil {
		t.Fatalf("NewNodeStore returned error: %v", err)
	}
	defer func() { ns.Close() }()

	if err := ns.Upsert(&mqttpkg.NodeInfo{ID: "id1", LongName: "first"}); err != nil {
		t.Fatalf("Upsert insert returned error: %v", err)
	}
	if err := ns.Upsert(&mqttpkg.NodeInfo{ID: "id2", LongName: "second"}); err != nil {
		t.Fatalf("Upsert insert returned error: %v", err)
	}

	if err := ns.Upsert(&mqttpkg.NodeInfo{ID: "id1", LongName: "first updated"}); err != nil {
		t.Fatalf("Upsert update returned error: %v", err)
	}

	nodes, err := ns.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	m := map[string]*mqttpkg.NodeInfo{}
	for _, n := range nodes {
		m[n.ID] = n
	}
	if m["id1"].LongName != "first updated" || m["id2"].LongName != "second" {
		t.Fatalf("nodes not stored correctly: %+v", m)
	}
}

func TestNodeStoreListOrder(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nodes.db")

	ns, err := NewNodeStore(dbPath)
	if err != nil {
		t.Fatalf("NewNodeStore returned error: %v", err)
	}
	defer func() { ns.Close() }()

	ids := []string{"c", "a", "b"}
	for _, id := range ids {
		if err := ns.Upsert(&mqttpkg.NodeInfo{ID: id}); err != nil {
			t.Fatalf("Upsert returned error: %v", err)
		}
	}

	nodes, err := ns.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(nodes) != len(want) {
		t.Fatalf("expected %d nodes, got %d", len(want), len(nodes))
	}
	for i, n := range nodes {
		if n.ID != want[i] {
			t.Fatalf("unexpected order: got %v want %v", n.ID, want[i])
		}
	}
}

func TestNodeStoreAddTelemetry(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nodes.db")

	ns, err := NewNodeStore(dbPath)
	if err != nil {
		t.Fatalf("NewNodeStore returned error: %v", err)
	}
	defer ns.Close()

	tel := &latestpb.Telemetry{
		Time: 42,
		Variant: &latestpb.Telemetry_DeviceMetrics{DeviceMetrics: &latestpb.DeviceMetrics{
			BatteryLevel: proto.Uint32(90),
			Voltage:      proto.Float32(3.7),
		}},
	}
	if err := ns.AddTelemetry(tel); err != nil {
		t.Fatalf("AddTelemetry returned error: %v", err)
	}

	recs, err := ns.Telemetry()
	if err != nil {
		t.Fatalf("Telemetry returned error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	r := recs[0]
	if r.BatteryLevel != 90 || r.Time != 42 || (r.Voltage < 3.69 || r.Voltage > 3.71) {
		t.Fatalf("unexpected record %+v", r)
	}
}
