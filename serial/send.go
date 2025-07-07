package serial

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// SendTextMessage sends a text message to the primary channel via meshtastic-go.
func SendTextMessage(port, text string) error {
	return SendTextMessageTo(port, "", text)
}

// SendTextMessageTo sends a text message via meshtastic-go to the specified
// destination node. If dest is empty the message is broadcast.
func SendTextMessageTo(port, dest, text string) error {
	args := []string{"--port", port, "message", "send", "-m", text}
	if dest != "" {
		args = append(args, "--dest", dest)
	}
	cmd := exec.Command("/usr/local/bin/meshtastic-go", args...)
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
