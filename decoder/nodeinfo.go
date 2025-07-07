package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

const (
	start1    = 0x94
	start2    = 0xC3
	start1v21 = 0x44
	start2v21 = 0x03
	headerLen = 4
)

// DecodeNodeInfo decodes a protobuf blob into a NodeInfo message.
// Currently only the "latest" proto version is supported.
func DecodeNodeInfo(data []byte, version string) (*latestpb.NodeInfo, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return nil, err
	}
	switch version {
	case "", "latest", "2.1":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil && fr.GetNodeInfo() != nil {
			return fr.GetNodeInfo(), nil
		}
		var ni latestpb.NodeInfo
		if err := proto.Unmarshal(data, &ni); err == nil && (ni.Num != 0 || ni.User != nil) {
			return &ni, nil
		}
		return nil, fmt.Errorf("not a NodeInfo message")
	default:
		return nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}
