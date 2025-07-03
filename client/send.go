package mqtt

import (
	"fmt"
	"os/exec"
)

// SendTextMessage sends a text message to the primary channel via meshtastic-go.
func SendTextMessage(port, text string) error {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "--sendtext", text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd error: %v - %s", err, string(output))
	}
	return nil
}
