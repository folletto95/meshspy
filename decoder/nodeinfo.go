package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

const (
	start1    = 0x94
	start2    = 0xC3
	headerLen = 4
)

// DecodeNodeInfo decodes a protobuf blob into a NodeInfo message.
// Currently only the "latest" proto version is supported.
func DecodeNodeInfo(data []byte, version string) (*latestpb.NodeInfo, error) {
	if len(data) >= headerLen && data[0] == start1 && data[1] == start2 {
		l := int(data[2])<<8 | int(data[3])
		if len(data) >= headerLen+l {
			data = data[headerLen : headerLen+l]
		} else {
			return nil, fmt.Errorf("incomplete frame")
		}
	}
	switch version {
	case "", "latest":
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
