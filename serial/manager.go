package serial

import (
	"fmt"
	"log"
	"sync"
	"time"

	seriallib "go.bug.st/serial"
	"google.golang.org/protobuf/proto"

	"meshspy/nodemap"
	latestpb "meshspy/proto/latest/meshtastic"
)

// Manager provides exclusive access to a serial port. It opens the
// port once and allows sending commands and starting a read loop
// without reopening the device.
type Manager struct {
	name string
	baud int
	port seriallib.Port
	mu   sync.Mutex
}

// OpenManager opens the given serial port at the specified baud rate
// and returns a Manager that can be used for reading and writing.
func OpenManager(portName string, baud int) (*Manager, error) {
	p, err := seriallib.Open(portName, &seriallib.Mode{BaudRate: baud})
	if err != nil {
		return nil, err
	}
	p.SetReadTimeout(5 * time.Second)
	return &Manager{name: portName, baud: baud, port: p}, nil
}

// Close closes the underlying serial port.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.port == nil {
		return nil
	}
	err := m.port.Close()
	m.port = nil
	return err
}

// Send writes the given string to the serial port using the existing handle.
func (m *Manager) Send(data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.port == nil {
		return fmt.Errorf("serial port not open")
	}
	log.Printf("\u2191 write to %s: %q", m.name, data)
	_, err := m.port.Write([]byte(data))
	return err
}

// SendTextMessage sends a broadcast text message over the mesh network
// using the open serial port.
func (m *Manager) SendTextMessage(text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.port == nil {
		return fmt.Errorf("serial port not open")
	}
	pkt := &latestpb.MeshPacket{
		To: 0xffffffff,
		PayloadVariant: &latestpb.MeshPacket_Decoded{
			Decoded: &latestpb.Data{
				Portnum: latestpb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(text),
			},
		},
	}
	tr := &latestpb.ToRadio{
		PayloadVariant: &latestpb.ToRadio_Packet{Packet: pkt},
	}
	payload, err := proto.Marshal(tr)
	if err != nil {
		return err
	}
	frame := make([]byte, 4+len(payload))
	frame[0] = 0x94
	frame[1] = 0xC3
	frame[2] = byte(len(payload) >> 8)
	frame[3] = byte(len(payload))
	copy(frame[4:], payload)
	log.Printf("\u2191 write text to %s: %q", m.name, text)
	_, err = m.port.Write(frame)
	return err
}

// ReadLoop starts reading from the serial port using the same logic as the
// standalone ReadLoop function, but without reopening the port.
func (m *Manager) ReadLoop(debug bool, protoVersion string, nm *nodemap.Map,
	handleNodeInfo func(*latestpb.NodeInfo),
	handleMyInfo func(*latestpb.MyNodeInfo),
	handleTelemetry func(*latestpb.Telemetry),
	handleText func(string),
	publish func(string)) {

	readLoop(m.port, m.name, m.baud, debug, protoVersion, nm,
		handleNodeInfo, handleMyInfo, handleTelemetry, handleText, publish)
}
