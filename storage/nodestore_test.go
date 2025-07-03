package storage

import (
	"os"
	"path/filepath"
	"testing"

	mqttpkg "meshspy/client"
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
