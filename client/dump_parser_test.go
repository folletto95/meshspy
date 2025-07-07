package mqtt

import "testing"

func TestParseNodeInfoDump(t *testing.T) {
	sample := `Node Info
Num 3186665544
User id:"!bdf0a848" long_name:"Kynny_4K" short_name:"Ky4K"
FirmwareVersion 2.7.0.705515a
HwModel TBEAM
Role CLIENT

Node Info
Num 1294755766
User id:"!4d2c67b6" long_name:"Nicco-Mob5 \xf0\x9f\x8f\x8e" short_name:"\xf0\x9f\x8f\x8e"
FirmwareVersion 2.1.0
HwModel TBEAM
Role CLIENT
`
	nodes := ParseNodeInfoDump([]byte(sample))
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	n0 := nodes[0]
	if n0.ID != "!bdf0a848" || n0.LongName != "Kynny_4K" || n0.ShortName != "Ky4K" || n0.FirmwareVersion != "2.7.0.705515a" {
		t.Fatalf("unexpected node0: %+v", n0)
	}
	n1 := nodes[1]
	if n1.ID != "!4d2c67b6" || n1.FirmwareVersion != "2.1.0" {
		t.Fatalf("unexpected node1: %+v", n1)
	}
}

func TestProtoVersionForFirmware(t *testing.T) {
	if v := ProtoVersionForFirmware("2.7.0.705515a"); v != "latest" {
		t.Fatalf("unexpected version %s", v)
	}
	if v := ProtoVersionForFirmware("2.1.0"); v != "2.1" {
		t.Fatalf("unexpected version for 2.1.x: %s", v)
	}
	if v := ProtoVersionForFirmware("2.1.5"); v != "2.1" {
		t.Fatalf("unexpected version for 2.1.x: %s", v)
	}
	if v := ProtoVersionForFirmware(""); v != "latest" {
		t.Fatalf("unexpected empty version %s", v)
	}
}
