package mgmtapi

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	latestpb "meshspy/proto/latest/meshtastic"
)

func TestSendTelemetry(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/telemetry" {
			t.Fatalf("path %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		body = string(b)
	}))
	defer srv.Close()

	c := New(srv.URL)
	tel := &latestpb.Telemetry{Time: 1}
	if err := c.SendTelemetry(tel); err != nil {
		t.Fatalf("send: %v", err)
	}
	if !strings.Contains(body, "\"time\":1") {
		t.Fatalf("unexpected body %s", body)
	}
}

func TestSendWaypoint(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
	}))
	defer srv.Close()
	c := New(srv.URL)
	wp := &latestpb.Waypoint{Name: "here"}
	if err := c.SendWaypoint(wp); err != nil {
		t.Fatalf("send: %v", err)
	}
	if path != "/api/waypoints" {
		t.Fatalf("path %s", path)
	}
}

func TestSendAdmin(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
	}))
	defer srv.Close()
	c := New(srv.URL)
	if err := c.SendAdmin([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if !strings.Contains(got, base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})) {
		t.Fatalf("bad body %s", got)
	}
}

func TestSendAlert(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
	}))
	defer srv.Close()
	c := New(srv.URL)
	if err := c.SendAlert("boom"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if !strings.Contains(got, "boom") {
		t.Fatalf("bad body %s", got)
	}
}
