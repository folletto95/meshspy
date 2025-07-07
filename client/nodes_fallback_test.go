package mqtt

import (
	"os"
	"testing"
)

func TestGetMeshNodesFallback(t *testing.T) {
	scriptPath := "/usr/local/bin/meshtastic-go"
	script := "#!/bin/sh\n" +
		"if [ \"$3\" = \"nodes\" ]; then\n" +
		"  echo 'No help topic for nodes' >&2\n" +
		"  exit 3\n" +
		"fi\n" +
		"echo '| 123 | NodeOne | N/A | 0 | 123456789 | 987654321 |'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create stub script: %v", err)
	}
	defer os.Remove(scriptPath)

	nodes, err := GetMeshNodes("ttyS1")
	if err != nil {
		t.Fatalf("GetMeshNodes returned error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].ID != "0x7b" || nodes[0].LongName != "NodeOne" {
		t.Fatalf("unexpected node: %+v", nodes[0])
	}
}
