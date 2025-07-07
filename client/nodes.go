package mqtt

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetMeshNodes retrieves the list of nodes from meshtastic-go using the 'nodes' command.
func GetMeshNodes(port string) ([]*NodeInfo, error) {
	cmd := exec.Command("/usr/local/bin/meshtastic-go", "--port", port, "nodes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ParseNodesOutput(output)
}

// ParseNodesOutput parses a table of nodes produced by meshtastic-go.
func ParseNodesOutput(data []byte) ([]*NodeInfo, error) {
	var nodes []*NodeInfo
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}
		idStr := strings.TrimSpace(fields[1])
		name := strings.TrimSpace(fields[2])
		snrStr := strings.TrimSpace(fields[3])
		latStr := strings.TrimSpace(fields[5])
		lonStr := strings.TrimSpace(fields[6])
		num, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			continue
		}
		var snr float64
		if snrStr != "" && snrStr != "N/A" {
			snr, _ = strconv.ParseFloat(snrStr, 64)
		}
		lat, _ := strconv.ParseInt(latStr, 10, 64)
		lon, _ := strconv.ParseInt(lonStr, 10, 64)
		nodes = append(nodes, &NodeInfo{
			ID:        fmt.Sprintf("0x%x", num),
			Num:       uint32(num),
			LongName:  name,
			Snr:       snr,
			Latitude:  float64(lat) / 1e7,
			Longitude: float64(lon) / 1e7,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}
