package serial

import (
	"os"
	"strings"
	"testing"
)

func TestSendTextMessage(t *testing.T) {
	scriptPath := "/usr/local/bin/meshtastic-go"
	script := "#!/bin/sh\necho \"$@\" > /tmp/sendtext_args\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create stub script: %v", err)
	}
	defer os.Remove(scriptPath)
	defer os.Remove("/tmp/sendtext_args")

	if err := SendTextMessage("ttyS1", "hello world"); err != nil {
		t.Fatalf("SendTextMessage returned error: %v", err)
	}
	data, err := os.ReadFile("/tmp/sendtext_args")
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	got := strings.TrimSpace(string(data))
	want := "--port ttyS1 message send -m hello world"
	if got != want {
		t.Fatalf("unexpected args: got %q want %q", got, want)
	}
}
