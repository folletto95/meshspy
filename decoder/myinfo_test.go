package decoder

import (
	"testing"

	"google.golang.org/protobuf/proto"
	pb "meshspy/proto/latest/meshtastic"
)

func TestDecodeMyInfo(t *testing.T) {
	orig := &pb.MyNodeInfo{MyNodeNum: 123, RebootCount: 2}
	data, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mi, err := DecodeMyInfo(data, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if mi.GetMyNodeNum() != 123 {
		t.Fatalf("unexpected node num %d", mi.GetMyNodeNum())
	}
}

func TestDecodeMyInfoFramed(t *testing.T) {
	orig := &pb.MyNodeInfo{MyNodeNum: 7, RebootCount: 1}
	fr := &pb.FromRadio{PayloadVariant: &pb.FromRadio_MyInfo{MyInfo: orig}}
	payload, err := proto.Marshal(fr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	header := []byte{0x94, 0xC3, byte(len(payload) >> 8), byte(len(payload))}
	frame := append(header, payload...)
	mi, err := DecodeMyInfo(frame, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if mi.GetMyNodeNum() != 7 {
		t.Fatalf("unexpected result %+v", mi)
	}
}
