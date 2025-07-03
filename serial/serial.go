package serial

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	serial "go.bug.st/serial"

	"meshspy/decoder"
	"meshspy/nodemap"
)

var nodeRe = regexp.MustCompile(`(?:from|fr|id)=(0x[0-9a-fA-F]+)`)
var fallbackRe = regexp.MustCompile(`(?:from|fr|id) (0x[0-9a-fA-F]+)`)
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ReadLoop apre la porta seriale e legge in loop, pubblicando i pacchetti validi
func ReadLoop(portName string, baud int, debug bool, nm *nodemap.Map, publish func(string)) {
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

		if nm != nil {
			if ni, err := decoder.DecodeNodeInfo([]byte(line), "latest"); err == nil {
				nm.UpdateFromProto(ni)
				if debug {
					log.Printf("[DEBUG nodemap] learned %s => %s/%s", fmt.Sprintf("0x%x", ni.GetNum()), ni.GetUser().GetLongName(), ni.GetUser().GetShortName())
				}
				continue
			}
		}

		node := parseNodeName(line)
		if node == "" || node == "0x0" {
			if debug {
				log.Printf("[DEBUG parse] no node found in %q", line)
			}
			continue
		}
		if nm != nil {
			node = nm.Resolve(node)
		}
		if node == lastNode {
			continue
		}
		lastNode = node

		payload := fmt.Sprintf(`{"node":"%s","ts":%d}`, node, time.Now().Unix())
		publish(payload)
	}
}

// cleanLine rimuove i codici ANSI da una riga
func cleanLine(line string) string {
	return ansiEscape.ReplaceAllString(line, "")
}

// parseNodeName estrae l'identificativo del nodo da una riga
func parseNodeName(line string) string {
	if m := nodeRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	if m := fallbackRe.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	return ""
}

// Send apre la porta seriale, invia i dati e chiude la porta.
func Send(portName string, baud int, data string) error {
	port, err := serial.Open(portName, &serial.Mode{BaudRate: baud})
	if err != nil {
		return err
	}
	defer port.Close()
	_, err = port.Write([]byte(data))
	return err
}
