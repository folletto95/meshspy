package mqtt

import (
	"math"
	"testing"

	pb "meshspy/proto/latest/meshtastic"
)

func TestNodeInfoFromProto(t *testing.T) {
	ni := &pb.NodeInfo{
		Num:                   0x12,
		User:                  &pb.User{Id: "id", LongName: "Alice", ShortName: "A"},
		Snr:                   7.5,
		LastHeard:             123,
		Channel:               1,
		ViaMqtt:               true,
		HopsAway:              protoUint32(2),
		IsFavorite:            true,
		IsIgnored:             true,
		IsKeyManuallyVerified: true,
	}
	info := NodeInfoFromProto(ni)
	if info == nil {
		t.Fatal("nil info")
	}
	if info.Num != 0x12 {
		t.Fatalf("num=%d", info.Num)
	}
	if info.Snr != 7.5 {
		t.Fatalf("snr=%v", info.Snr)
	}
	if info.LastHeard != 123 {
		t.Fatalf("lastHeard=%v", info.LastHeard)
	}
	if info.Channel != 1 || !info.ViaMqtt || info.HopsAway != 2 {
		t.Fatalf("channel=%d via=%v hops=%d", info.Channel, info.ViaMqtt, info.HopsAway)
	}
	if !info.IsFavorite || !info.IsIgnored || !info.IsKeyManuallyVerified {
		t.Fatalf("bool flags incorrect")
	}
}

func TestNodeInfoFromProtoMetrics(t *testing.T) {
	ni := &pb.NodeInfo{
		User: &pb.User{Id: "id"},
		DeviceMetrics: &pb.DeviceMetrics{
			BatteryLevel:       protoUint32(80),
			Voltage:            protoFloat32(3.3),
			ChannelUtilization: protoFloat32(4.5),
			AirUtilTx:          protoFloat32(2.1),
			UptimeSeconds:      protoUint32(3600),
		},
		Position: &pb.Position{
			LatitudeI:      protoInt32(123456789),
			LongitudeI:     protoInt32(-123456789),
			Altitude:       protoInt32(250),
			Time:           42,
			LocationSource: pb.Position_LOC_INTERNAL,
		},
	}
	info := NodeInfoFromProto(ni)
	if info == nil {
		t.Fatal("nil info")
	}
	if info.BatteryLevel != 80 {
		t.Fatalf("battery=%d", info.BatteryLevel)
	}
	if math.Abs(info.Voltage-3.3) > 1e-6 {
		t.Fatalf("voltage=%v", info.Voltage)
	}
	if math.Abs(info.ChannelUtil-4.5) > 1e-6 || math.Abs(info.AirUtilTx-2.1) > 1e-6 || info.UptimeSeconds != 3600 {
		t.Fatalf("util mismatch: %+v", info)
	}
	if info.Latitude != float64(123456789)/1e7 || info.Longitude != float64(-123456789)/1e7 {
		t.Fatalf("position mismatch: %+v", info)
	}
	if info.LocationSource != pb.Position_LOC_INTERNAL.String() {
		t.Fatalf("location source mismatch: %s", info.LocationSource)
	}
}

func protoFloat32(v float32) *float32 { return &v }
func protoInt32(v int32) *int32       { return &v }

func protoUint32(v uint32) *uint32 { return &v }
