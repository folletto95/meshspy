package serial

import "testing"

func TestParseNodeName(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"[INF] packet from=0x1a2b seq=1", "0x1a2b"},
		{"[DBG] other from 0xABCD tail", "0xABCD"},
		{"node status from=0x0", "0x0"},
		{"prefix from=0xDEAD data from 0xBEEF", "0xDEAD"},
		{"garbage log line", ""},
		{"from=nothex", ""},
	}
	for _, tt := range tests {
		if got := parseNodeName(tt.line); got != tt.want {
			t.Errorf("parseNodeName(%q) = %q; want %q", tt.line, got, tt.want)
		}
	}
}