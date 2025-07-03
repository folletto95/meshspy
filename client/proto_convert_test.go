package mqtt

import (
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

func protoUint32(v uint32) *uint32 { return &v }
