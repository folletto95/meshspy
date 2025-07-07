package serial

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	serial "go.bug.st/serial"

	"meshspy/decoder"
	"meshspy/nodemap"
	latestpb "meshspy/proto/latest/meshtastic"
)

var nodeRe = regexp.MustCompile(`(?:from|fr|id)=(0x[0-9a-fA-F]+)`)
var fallbackRe = regexp.MustCompile(`(?:from|fr|id) (0x[0-9a-fA-F]+)`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ReadLoop opens the serial port and decodes incoming protobuf messages.
// It invokes the provided callbacks for NodeInfo, Telemetry and text messages.
// It also publishes the identifiers of detected nodes using the publish function.
func ReadLoop(portName string, baud int, debug bool, protoVersion string, nm *nodemap.Map,
	handleNodeInfo func(*latestpb.NodeInfo),
	handleMyInfo func(*latestpb.MyNodeInfo),
	handleTelemetry func(*latestpb.Telemetry),
	handleText func(string),
	publish func(string)) {
	var (
		port serial.Port
		err  error
	)
	for i := 0; i < 5; i++ {
		port, err = serial.Open(portName, &serial.Mode{BaudRate: baud})
		if err == nil {
			port.SetReadTimeout(5 * time.Second)
			break
		}
		log.Printf("Failed to open serial port %s: %v (attempt %d/5)", portName, err, i+1)
		time.Sleep(time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to open serial port %s after 5 attempts: %v", portName, err)
	}
	defer port.Close()

	readLoop(port, portName, baud, debug, protoVersion, nm,
		handleNodeInfo, handleMyInfo, handleTelemetry, handleText, publish)
}

func readLoop(port serial.Port, portName string, baud int, debug bool, protoVersion string, nm *nodemap.Map,
	handleNodeInfo func(*latestpb.NodeInfo),
	handleMyInfo func(*latestpb.MyNodeInfo),
	handleTelemetry func(*latestpb.Telemetry),
	handleText func(string),
	publish func(string)) {
	log.Printf("Listening on serial %s at %d baud", portName, baud)

	const (
		start1    = 0x94
		start2    = 0xC3
		headerLen = 4
		maxSize   = 512
	)

	var (
		buf      []byte
		logBuf   bytes.Buffer
		lastNode string
	)

	handleLine := func(line string) {
		line = cleanLine(line)
		if debug {
			log.Printf("[DEBUG serial] %q", line)
		}

		if nm != nil {
			if ni, err := decoder.DecodeNodeInfo([]byte(line), protoVersion); err == nil {
				nm.UpdateFromProto(ni)
				if handleNodeInfo != nil {
					handleNodeInfo(ni)
				}
				if debug {
					log.Printf("[DEBUG nodemap] learned %s => %s/%s", fmt.Sprintf("0x%x", ni.GetNum()), ni.GetUser().GetLongName(), ni.GetUser().GetShortName())
				}
				return
			}
		}
		if mi, err := decoder.DecodeMyInfo([]byte(line), protoVersion); err == nil {
			if handleMyInfo != nil {
				handleMyInfo(mi)
			}
			return
		}
		if tel, err := decoder.DecodeTelemetry([]byte(line), protoVersion); err == nil {
			if handleTelemetry != nil {
				handleTelemetry(tel)
			}
			return
		}

		if txt, err := decoder.DecodeText([]byte(line), protoVersion); err == nil {
			if handleText != nil {
				handleText(txt)
			}
			return
		}

		node := parseNodeName(line)
		if node == "" || node == "0x0" {
			if debug {
				log.Printf("[DEBUG parse] no node found in %q", line)
			}
			return
		}
		if nm != nil {
			node = nm.ResolveLong(node)
		}
		if node == lastNode {
			return
		}
		lastNode = node

		payload := fmt.Sprintf(`{"node":"%s"}`, node)
		publish(payload)
	}

	for {
		var b [1]byte
		n, err := port.Read(b[:])
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Printf("Serial read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if n == 0 {
			continue
		}
		buf = append(buf, b[0])

		for len(buf) > 0 {
			if len(buf) < 2 {
				break
			}
			if buf[0] != start1 || buf[1] != start2 {
				ch := buf[0]
				buf = buf[1:]
				if ch == '\n' {
					handleLine(logBuf.String())
					logBuf.Reset()
				} else if ch != '\r' {
					logBuf.WriteByte(ch)
				}
				continue
			}
			if len(buf) < headerLen {
				break
			}
			length := int(buf[2])<<8 | int(buf[3])
			if length <= 0 || length > maxSize {
				buf = buf[1:]
				continue
			}
			if len(buf) < headerLen+length {
				break
			}
			payload := buf[headerLen : headerLen+length]
			if nm != nil {
				if ni, err := decoder.DecodeNodeInfo(payload, protoVersion); err == nil {
					nm.UpdateFromProto(ni)
					if handleNodeInfo != nil {
						handleNodeInfo(ni)
					}
					if debug {
						log.Printf("[DEBUG nodemap] learned %s => %s/%s", fmt.Sprintf("0x%x", ni.GetNum()), ni.GetUser().GetLongName(), ni.GetUser().GetShortName())
					}
				}
			}
			if mi, err := decoder.DecodeMyInfo(payload, protoVersion); err == nil {
				if handleMyInfo != nil {
					handleMyInfo(mi)
				}
			}

			if txt, err := decoder.DecodeText(payload, protoVersion); err == nil {
				if handleText != nil {
					handleText(txt)
				}
			} else if tele, err := decoder.DecodeTelemetry(payload, protoVersion); err == nil {
				if handleTelemetry != nil {
					handleTelemetry(tele)
				}
			}
			buf = buf[headerLen+length:]
		}
	}
}

// cleanLine removes ANSI escape codes from a line
func cleanLine(line string) string {
	return ansiEscape.ReplaceAllString(line, "")
}

// parseNodeName extracts the node identifier from a line
func parseNodeName(line string) string {
	if m := nodeRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	if m := fallbackRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	return ""
}

// Send opens the serial port, writes the data and closes the port.
func Send(portName string, baud int, data string) error {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baud})
	if err != nil {
		return err
	}
	defer port.Close()
	log.Printf("\u2191 write to %s: %q", portName, data)
	_, err = port.Write([]byte(data))
	return err
}
