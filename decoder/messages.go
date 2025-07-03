package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

// stripFrame removes the radio framing bytes if present.
func stripFrame(data []byte) ([]byte, error) {
	if len(data) >= headerLen && data[0] == start1 && data[1] == start2 {
		l := int(data[2])<<8 | int(data[3])
		if len(data) >= headerLen+l {
			return data[headerLen : headerLen+l], nil
		}
		return nil, fmt.Errorf("incomplete frame")
	}
	return data, nil
}

// DecodeTelemetry decodes a protobuf Telemetry message from the given data.
// It supports the same framing as DecodeNodeInfo.
func DecodeTelemetry(data []byte, version string) (*latestpb.Telemetry, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return nil, err
	}
	switch version {
	case "", "latest":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_TELEMETRY_APP {
						var tm latestpb.Telemetry
						if err := proto.Unmarshal(dec.GetPayload(), &tm); err == nil {
							return &tm, nil
						}
					}
				}
			}
		}
		var tm latestpb.Telemetry
		if err := proto.Unmarshal(data, &tm); err == nil && tm.GetVariant() != nil {
			return &tm, nil
		}
		return nil, fmt.Errorf("not a Telemetry message")
	default:
		return nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}

// DecodeText extracts a plain text message from the given data.
func DecodeText(data []byte, version string) (string, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return "", err
	}
	switch version {
	case "", "latest":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_TEXT_MESSAGE_APP {
						return string(dec.GetPayload()), nil
					}
				}
			}
		}
		var d latestpb.Data
		if err := proto.Unmarshal(data, &d); err == nil && d.GetPortnum() == latestpb.PortNum_TEXT_MESSAGE_APP {
			return string(d.GetPayload()), nil
		}
		return "", fmt.Errorf("not a text message")
	default:
		return "", fmt.Errorf("unsupported proto version: %s", version)
	}
}
