package decoder

import (
	"testing"

	"google.golang.org/protobuf/proto"
	pb "meshspy/proto/latest/meshtastic"
)

func TestDecodeTelemetry(t *testing.T) {
	orig := &pb.Telemetry{
		Time:    1234,
		Variant: &pb.Telemetry_DeviceMetrics{DeviceMetrics: &pb.DeviceMetrics{BatteryLevel: proto.Uint32(80)}},
	}
	data, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	tel, err := DecodeTelemetry(data, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if tel.GetTime() != 1234 {
		t.Fatalf("unexpected time %d", tel.GetTime())
	}
}
