package serial

import (
	"os/exec"
)

// SendText uses meshtastic-go to send a text message over the mesh network.
// It executes: meshtastic-go --port <port> --sendtext <msg>.
func SendText(port, msg string) error {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "--sendtext", msg)
	return cmd.Run()
}