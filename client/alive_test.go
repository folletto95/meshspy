package mqtt

import (
	"errors"
	"testing"
	"time"

	mqttpkg "github.com/eclipse/paho.mqtt.golang"
	"meshspy/config"
)

// mockToken is a minimal implementation of mqtt.Token.
type mockToken struct{ err error }

func (t *mockToken) Wait() bool                       { return true }
func (t *mockToken) WaitTimeout(_ time.Duration) bool { return true }
func (t *mockToken) Done() <-chan struct{}            { ch := make(chan struct{}); close(ch); return ch }
func (t *mockToken) Error() error                     { return t.err }

// mockClient implements mqtt.Client for testing PublishAlive.
type mockClient struct {
	payloads [][]byte
	err      error
}

func (m *mockClient) IsConnected() bool      { return true }
func (m *mockClient) IsConnectionOpen() bool { return true }
func (m *mockClient) Connect() mqttpkg.Token { return &mockToken{} }
func (m *mockClient) Disconnect(uint)        {}
func (m *mockClient) Subscribe(string, byte, mqttpkg.MessageHandler) mqttpkg.Token {
	return &mockToken{}
}
func (m *mockClient) SubscribeMultiple(map[string]byte, mqttpkg.MessageHandler) mqttpkg.Token {
	return &mockToken{}
}
func (m *mockClient) Unsubscribe(...string) mqttpkg.Token     { return &mockToken{} }
func (m *mockClient) AddRoute(string, mqttpkg.MessageHandler) {}
func (m *mockClient) OptionsReader() mqttpkg.ClientOptionsReader {
	opts := mqttpkg.NewClientOptions()
	return mqttpkg.NewOptionsReader(opts)
}
func (m *mockClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqttpkg.Token {
	var b []byte
	switch v := payload.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	}
	m.payloads = append(m.payloads, b)
	return &mockToken{err: m.err}
}

func TestSendAliveIfNeeded(t *testing.T) {
	cfg := config.Config{SendAlive: true, MQTTTopic: "test"}
	mc := &mockClient{}
	if err := SendAliveIfNeeded(mc, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mc.payloads) != 1 {
		t.Fatalf("publish calls=%d", len(mc.payloads))
	}
	if string(mc.payloads[0]) != "MeshSpy Alive" {
		t.Fatalf("payload=%q", mc.payloads[0])
	}
}

func TestSendAliveIfNeededError(t *testing.T) {
	cfg := config.Config{SendAlive: true, MQTTTopic: "test"}
	mc := &mockClient{err: errors.New("fail")}
	if err := SendAliveIfNeeded(mc, cfg); err == nil {
		t.Fatalf("expected error")
	}
}
