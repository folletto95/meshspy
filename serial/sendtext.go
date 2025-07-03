package serial

import (
	"log"
	"os/exec"
	"strings"
)

// SendText uses meshtastic-go to send a text message over the mesh network.
// It executes: meshtastic-go --port <port> --sendtext <msg>.
func SendText(port, msg string) error {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "--sendtext", msg)
	log.Printf("\u2191 sending command: %s", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("command output: %s", string(output))
		return err
	}
	if len(output) > 0 {
		log.Printf("command output: %s", string(output))
	}
	return nil
}
