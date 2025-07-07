package mgmtapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	mqttpkg "meshspy/client"
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
