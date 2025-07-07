package decoder

import (
	"testing"

	"google.golang.org/protobuf/proto"
	pb "meshspy/proto/latest/meshtastic"
)

func TestDecodeNodeInfo(t *testing.T) {
	orig := &pb.NodeInfo{
		Num:  42,
		User: &pb.User{Id: "abc", LongName: "Alice", ShortName: "A"},
	}
	data, err := proto.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	ni, err := DecodeNodeInfo(data, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if ni.GetUser().GetId() != "abc" {
		t.Fatalf("unexpected id %q", ni.GetUser().GetId())
	}
}

func TestDecodeNodeInfoFramed(t *testing.T) {
	orig := &pb.NodeInfo{
		Num:  77,
		User: &pb.User{Id: "xyz", LongName: "Bob", ShortName: "B"},
	}
	fr := &pb.FromRadio{PayloadVariant: &pb.FromRadio_NodeInfo{NodeInfo: orig}}
	payload, err := proto.Marshal(fr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	header := []byte{0x94, 0xC3, byte(len(payload) >> 8), byte(len(payload))}
	frame := append(header, payload...)
	ni, err := DecodeNodeInfo(frame, "latest")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if ni.GetNum() != orig.GetNum() || ni.GetUser().GetId() != "xyz" {
		t.Fatalf("unexpected result %+v", ni)
	}
}

func TestDecodeNodeInfoFramedV21(t *testing.T) {
	orig := &pb.NodeInfo{
		Num:  55,
		User: &pb.User{Id: "v21", LongName: "Charlie", ShortName: "C"},
	}
	fr := &pb.FromRadio{PayloadVariant: &pb.FromRadio_NodeInfo{NodeInfo: orig}}
	payload, err := proto.Marshal(fr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	header := []byte{0x44, 0x03, byte(len(payload) >> 8), byte(len(payload))}
	frame := append(header, payload...)
	ni, err := DecodeNodeInfo(frame, "2.1")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if ni.GetNum() != orig.GetNum() || ni.GetUser().GetId() != "v21" {
		t.Fatalf("unexpected result %+v", ni)
	}
}
