package mqtt

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestGetLocalNodeInfoScanError(t *testing.T) {
	scriptPath := "/usr/local/bin/meshtastic-go"
	// create stub script
	script := "#!/bin/sh\npython3 - <<'PY'\nprint('Node Info')\nprint('a'*70000)\nPY\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create stub script: %v", err)
	}
	defer os.Remove(scriptPath)

	_, err := GetLocalNodeInfo("dummy")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected ErrTooLong, got %v", err)
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("script did not run correctly: %v", err)
	}
}

func TestLoadNodeInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.json")
	want := &NodeInfo{LongName: "Foo", FirmwareVersion: "1.0"}
	if err := SaveNodeInfo(want, path); err != nil {
		t.Fatalf("SaveNodeInfo failed: %v", err)
	}
	got, err := LoadNodeInfo(path)
	if err != nil {
		t.Fatalf("LoadNodeInfo returned error: %v", err)
	}
	if got.LongName != want.LongName || got.FirmwareVersion != want.FirmwareVersion {
		t.Fatalf("mismatch: got %+v want %+v", got, want)
	}
}

func TestGetLocalNodeInfoCachedUsesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.json")
	want := &NodeInfo{LongName: "Bar", FirmwareVersion: "2.0"}
	if err := SaveNodeInfo(want, path); err != nil {
		t.Fatalf("SaveNodeInfo failed: %v", err)
	}

	scriptPath := "/usr/local/bin/meshtastic-go"
	marker := filepath.Join(dir, "executed")
	script := "#!/bin/sh\ntouch " + marker + "\nexit 1\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create stub script: %v", err)
	}
	defer os.Remove(scriptPath)

	got, err := GetLocalNodeInfoCached("dummy", path)
	if err != nil {
		t.Fatalf("GetLocalNodeInfoCached returned error: %v", err)
	}
	if got.LongName != want.LongName || got.FirmwareVersion != want.FirmwareVersion {
		t.Fatalf("unexpected info: got %+v want %+v", got, want)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("meshtastic-go was executed")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat marker: %v", err)
	}
}

func TestGetLocalNodeInfoCachedFetches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.json")

	scriptPath := "/usr/local/bin/meshtastic-go"
	script := "#!/bin/sh\necho 'Node Info'\necho 'User id:\"id1\" long_name:\"Baz\" short_name:\"B\" macaddr:\"12\" hw_model:TBEAM role:CLIENT'\necho 'FirmwareVersion 3.0'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create stub script: %v", err)
	}
	defer os.Remove(scriptPath)

	info, err := GetLocalNodeInfoCached("dummy", path)
	if err != nil {
		t.Fatalf("GetLocalNodeInfoCached returned error: %v", err)
	}
	if info.LongName != "Baz" || info.FirmwareVersion != "3.0" {
		t.Fatalf("unexpected info: %+v", info)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("info not saved: %v", err)
	}
}
