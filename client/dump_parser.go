package mqtt

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseNodeInfoDump parses a text dump produced by meshtastic-go info
// (or serial logs) and extracts NodeInfo entries found in it.
// It supports the minimal subset of fields needed to identify nodes
// and their firmware version.
func ParseNodeInfoDump(data []byte) []*NodeInfo {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var nodes []*NodeInfo
	var node *NodeInfo
	// regular expressions for parsing
	userRe := regexp.MustCompile(`long_name:"([^"]+)"\s+short_name:"([^"]+)"`)
	idRe := regexp.MustCompile(`id:"([^"]+)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "Node Info":
			if node != nil {
				nodes = append(nodes, node)
			}
			node = &NodeInfo{}
		case node != nil && strings.HasPrefix(line, "Num"):
			fields := strings.Fields(line)
			if len(fields) > 1 {
				if n, err := strconv.ParseUint(fields[1], 10, 32); err == nil {
					node.Num = uint32(n)
					node.ID = fmt.Sprintf("0x%x", node.Num)
				}
			}
		case node != nil && strings.HasPrefix(line, "User"):
			if m := userRe.FindStringSubmatch(line); len(m) == 3 {
				node.LongName = m[1]
				node.ShortName = m[2]
			}
			if m := idRe.FindStringSubmatch(line); len(m) == 2 {
				node.ID = m[1]
			}
		case node != nil && strings.HasPrefix(line, "FirmwareVersion"):
			node.FirmwareVersion = strings.TrimSpace(strings.TrimPrefix(line, "FirmwareVersion"))
		case node != nil && strings.HasPrefix(line, "HwModel"):
			node.RadioHwModel = strings.TrimSpace(strings.TrimPrefix(line, "HwModel"))
		case node != nil && strings.HasPrefix(line, "Role") && node.RadioRole == "":
			node.RadioRole = strings.TrimSpace(strings.TrimPrefix(line, "Role"))
		case node != nil && line == "":
			// blank line might indicate end of block
		case node != nil && strings.HasSuffix(line, "Settings"):
			// reached the next section
			nodes = append(nodes, node)
			node = nil
		}
	}
	if node != nil {
		nodes = append(nodes, node)
	}
	return nodes
}
