package mgmtapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	mqttpkg "meshspy/client"
	"meshspy/storage"

	"google.golang.org/protobuf/encoding/protojson"
	latestpb "meshspy/proto/latest/meshtastic"
)

// Client communicates with the management server HTTP API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a new API client for the given base URL. If url is empty,
// the returned client will be nil.
func New(url string) *Client {
	if url == "" {
		return nil
	}
	return &Client{
		baseURL: strings.TrimRight(url, "/"),
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// SendNode uploads a NodeInfo to the management server.
func (c *Client) SendNode(info *mqttpkg.NodeInfo) error {
	if c == nil {
		return nil
	}
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/nodes", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

// SendCommand sends a command string to the server which will publish it on MQTT.
func (c *Client) SendCommand(cmd string) error {
	if c == nil {
		return nil
	}
	payload := struct {
		Cmd string `json:"cmd"`
	}{Cmd: cmd}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/send", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

// ListNodes retrieves the known nodes from the server.
func (c *Client) ListNodes() ([]*mqttpkg.NodeInfo, error) {
	if c == nil {
		return nil, nil
	}
	resp, err := c.http.Get(c.baseURL + "/api/nodes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}
	var nodes []*mqttpkg.NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListPositions retrieves node positions from the server.
func (c *Client) ListPositions(nodeID string) ([]storage.NodePosition, error) {
	if c == nil {
		return nil, nil
	}
	url := c.baseURL + "/api/positions"
	if nodeID != "" {
		url += "?node=" + nodeID
	}
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}
	var pos []storage.NodePosition
	if err := json.NewDecoder(resp.Body).Decode(&pos); err != nil {
		return nil, err
	}
	return pos, nil
}

// SendTelemetry uploads a Telemetry message to the management server.
func (c *Client) SendTelemetry(t *latestpb.Telemetry) error {
	if c == nil || t == nil {
		return nil
	}
	b, err := protojson.Marshal(t)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/telemetry", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

// SendWaypoint uploads a Waypoint message to the management server.
func (c *Client) SendWaypoint(wp *latestpb.Waypoint) error {
	if c == nil || wp == nil {
		return nil
	}
	b, err := protojson.Marshal(wp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/waypoints", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

// SendAdmin uploads raw admin payload to the management server encoded as base64.
func (c *Client) SendAdmin(payload []byte) error {
	if c == nil || payload == nil {
		return nil
	}
	data := struct {
		Payload string `json:"payload"`
	}{Payload: base64.StdEncoding.EncodeToString(payload)}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/admin", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

// SendAlert uploads an alert text message to the management server.
func (c *Client) SendAlert(text string) error {
	if c == nil || text == "" {
		return nil
	}
	data := struct {
		Text string `json:"text"`
	}{Text: text}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/alerts", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}
