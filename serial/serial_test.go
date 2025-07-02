package serial

import "testing"

func TestParseNodeName(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"[INF] packet from=0x1a2b seq", "0x1a2b"},
		{"random text from 0xABCD end", "0xABCD"},
		{"node info from=0x0", "0x0"},
		{"no node here", ""},
	}
	for _, tt := range tests {
		if got := parseNodeName(tt.line); got != tt.want {
			t.Errorf("parseNodeName(%q) = %q; want %q", tt.line, got, tt.want)
		}
	}
}