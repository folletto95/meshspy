package serial

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	"github.com/tarm/serial"
)

var nodeRe = regexp.MustCompile(`from=(0x[0-9a-fA-F]+)`)
var fallbackRe = regexp.MustCompile(`from (0x[0-9a-fA-F]+)`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func ReadLoop(portName string, baud int, debug bool, publish func(string)) {
	cfg := &serial.Config{Name: portName, Baud: baud, ReadTimeout: time.Second * 5}
	port, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatalf("Failed to open serial port %s: %v", portName, err)
	}
	defer port.Close()

	reader := bufio.NewReader(port)
	log.Printf("Listening on serial %s at %d baud", portName, baud)

	var lastNode string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Printf("Serial read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		line = cleanLine(line)
		if debug {
			log.Printf("[DEBUG serial] %q", line)
		}

		node := parseNodeName(line)
		if node == "" || node == "0x0" {
			if debug {
				log.Printf("[DEBUG parse] no node found in %q", line)
			}
			continue
		}
		if node == lastNode {
			continue
		}
		lastNode = node

		payload := fmt.Sprintf(`{"node":"%s","ts":%d}`, node, time.Now().Unix())
		publish(payload)
	}
}

func cleanLine(line string) string {
	return ansiEscape.ReplaceAllString(line, "")
}

func parseNodeName(line string) string {
	if m := nodeRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	if m := fallbackRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	return ""
}
