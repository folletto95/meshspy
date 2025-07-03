package serial

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// SendTextMessage sends a text message to the primary channel via meshtastic-go.
func SendTextMessage(port, text string) error {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "--sendtext", text)
	log.Printf("\u2191 sending command: %s", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command output: %s", string(output))
		return fmt.Errorf("cmd error: %v", err)
	}
	if len(output) > 0 {
		log.Printf("command output: %s", string(output))
	}
	return nil
}
