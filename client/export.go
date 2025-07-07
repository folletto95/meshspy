package mqtt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ExportConfig runs meshtastic-go to export the configuration and save it to dest.
func ExportConfig(port, dest string) error {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "config")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd error: %v - %s", err, string(output))
	}
	return os.WriteFile(dest, output, 0644)
}

// BuildConfigFilename returns the backup file name.
func BuildConfigFilename(info *NodeInfo) string {
	sanitize := func(s string) string {
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "/", "-")
		return s
	}
	date := time.Now().Format("20060102")
	return fmt.Sprintf("%s-%s--%s-%s.txt", sanitize(info.LongName), sanitize(info.ShortName), info.FirmwareVersion, date)
}
