package mqtt

import "testing"

func TestParseNodesOutput(t *testing.T) {
	sample := "| 1773582993     | HB9ODI-Scereda 900m| N/A            | " +
		"950            | 458915999      | 89875399       |\n"
	nodes := ParseNodesOutput([]byte(sample))
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.ID != "0x69b6ba91" || n.Num != 1773582993 {
		t.Fatalf("id mismatch: %+v", n)
	}
	if n.LongName != "HB9ODI-Scereda 900m" {
		t.Fatalf("long name mismatch: %s", n.LongName)
	}
	if n.Snr != 0 {
		t.Fatalf("expected snr 0, got %v", n.Snr)
	}
	if n.Latitude != 45.8915999 || n.Longitude != 8.9875399 {
		t.Fatalf("position mismatch: %v,%v", n.Latitude, n.Longitude)
	}
}
