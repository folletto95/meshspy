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
