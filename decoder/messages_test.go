package decoder

import (
	"testing"

	"google.golang.org/protobuf/proto"
	pb "meshspy/proto/latest/meshtastic"
)

func TestDecodeText(t *testing.T) {
	d := &pb.Data{
		Portnum: pb.PortNum_TEXT_MESSAGE_APP,
		Payload: []byte("hello"),
	}
	mp := &pb.MeshPacket{PayloadVariant: &pb.MeshPacket_Decoded{Decoded: d}}
	fr := &pb.FromRadio{PayloadVariant: &pb.FromRadio_Packet{Packet: mp}}
	data, err := proto.Marshal(fr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	txt, err := DecodeText(data, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if txt != "hello" {
		t.Fatalf("unexpected text %q", txt)
	}
}

func TestDecodeTelemetry(t *testing.T) {
	tm := &pb.Telemetry{
		Time:    12345,
		Variant: &pb.Telemetry_DeviceMetrics{DeviceMetrics: &pb.DeviceMetrics{}},
	}
	payload, err := proto.Marshal(tm)
	if err != nil {
		t.Fatalf("marshal telemetry: %v", err)
	}
	d := &pb.Data{
		Portnum: pb.PortNum_TELEMETRY_APP,
		Payload: payload,
	}
	mp := &pb.MeshPacket{PayloadVariant: &pb.MeshPacket_Decoded{Decoded: d}}
	fr := &pb.FromRadio{PayloadVariant: &pb.FromRadio_Packet{Packet: mp}}
	data, err := proto.Marshal(fr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	dec, err := DecodeTelemetry(data, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if dec.GetTime() != tm.GetTime() {
		t.Fatalf("unexpected time %d", dec.GetTime())
	}
}
