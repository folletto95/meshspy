package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

// DecodeNodeInfo decodes a protobuf blob into a NodeInfo message.
// Currently only the "latest" proto version is supported.
func DecodeNodeInfo(data []byte, version string) (*latestpb.NodeInfo, error) {
	switch version {
	case "", "latest":
		var ni latestpb.NodeInfo
		if err := proto.Unmarshal(data, &ni); err == nil {
			return &ni, nil
		}
		// try wrapped in FromRadio
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil && fr.GetNodeInfo() != nil {
			return fr.GetNodeInfo(), nil
		}
		return nil, fmt.Errorf("not a NodeInfo message")
	default:
		return nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}
