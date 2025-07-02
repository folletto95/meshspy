package mqtt

import (
        "bufio"
        "errors"
        "io/fs"
        "os"
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