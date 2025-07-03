package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

// DecodeTelemetry decodes a protobuf blob into a Telemetry message.
func DecodeTelemetry(data []byte, version string) (*latestpb.Telemetry, error) {
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
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				if d := pkt.GetDecoded(); d != nil && d.GetPortnum() == latestpb.PortNum_TELEMETRY_APP {
					var tel latestpb.Telemetry
					if err := proto.Unmarshal(d.GetPayload(), &tel); err == nil {
						return &tel, nil
					}
				}
			}
		}
		var tel latestpb.Telemetry
		if err := proto.Unmarshal(data, &tel); err == nil && tel.Time != 0 {
			return &tel, nil
		}
		var d latestpb.Data
		if err := proto.Unmarshal(data, &d); err == nil && len(d.GetPayload()) > 0 && d.GetPortnum() == latestpb.PortNum_TELEMETRY_APP {
			if err := proto.Unmarshal(d.GetPayload(), &tel); err == nil {
				return &tel, nil
			}
		}
		return nil, fmt.Errorf("not a Telemetry message")
	default:
		return nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}
