package meshtastic

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type Info struct {
	NodeName string
	Firmware string
}

var nameRe = regexp.MustCompile(`(?i)Owner: (.+)`)
var fwRe = regexp.MustCompile(`(?i)Firmware: ([^\s]+)`)

func GetInfo(port string) (*Info, error) {
	cmd := exec.Command("meshtastic-go", "--port", port, "info")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(&out)
	info := &Info{}
	for scanner.Scan() {
		line := scanner.Text()
		if m := nameRe.FindStringSubmatch(line); len(m) == 2 {
			info.NodeName = strings.TrimSpace(m[1])
		}
		if m := fwRe.FindStringSubmatch(line); len(m) == 2 {
			info.Firmware = strings.TrimSpace(m[1])
		}
	}

	return info, nil
}
