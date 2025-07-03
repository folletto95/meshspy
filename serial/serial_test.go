package serial

import "testing"

func TestParseNodeName(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"[INF] packet from=0x1a2b seq=1", "0x1a2b"},
		{"[INF] packet fr=0x1234 seq=1", "0x1234"},
		{"[INF] packet id=0x4321", "0x4321"},
		{"[DBG] other from 0xABCD tail", "0xABCD"},
		{"[DBG] other fr 0x99AA tail", "0x99AA"},
		{"node status from=0x0", "0x0"},
		{"prefix id=0xDEAD data fr=0xBEEF", "0xDEAD"},
		{"garbage log line", ""},
		{"from=nothex", ""},
	}
	for _, tt := range tests {
		if got := parseNodeName(tt.line); got != tt.want {
			t.Errorf("parseNodeName(%q) = %q; want %q", tt.line, got, tt.want)
		}
	}
}
