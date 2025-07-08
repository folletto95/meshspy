package decoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	latestpb "meshspy/proto/latest/meshtastic"
)

// stripFrame removes the radio framing bytes if present.
func stripFrame(data []byte) ([]byte, error) {
	if len(data) >= headerLen {
		if (data[0] == start1 && data[1] == start2) ||
			(data[0] == start1v21 && data[1] == start2v21) {
			l := int(data[2])<<8 | int(data[3])
			if len(data) >= headerLen+l {
				return data[headerLen : headerLen+l], nil
			}
			return nil, fmt.Errorf("incomplete frame")
		}
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
	case "", "latest", "2.1":
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

// DecodeTelemetryWithID decodes a telemetry packet and returns the sender node
// id along with the Telemetry message. The id is zero when not present.
func DecodeTelemetryWithID(data []byte, version string) (uint32, *latestpb.Telemetry, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return 0, nil, err
	}
	switch version {
	case "", "latest", "2.1":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				from := pkt.GetFrom()
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_TELEMETRY_APP {
						var tm latestpb.Telemetry
						if err := proto.Unmarshal(dec.GetPayload(), &tm); err == nil {
							return from, &tm, nil
						}
					}
				}
			}
		}
		var tm latestpb.Telemetry
		if err := proto.Unmarshal(data, &tm); err == nil && tm.GetVariant() != nil {
			return 0, &tm, nil
		}
		return 0, nil, fmt.Errorf("not a Telemetry message")
	default:
		return 0, nil, fmt.Errorf("unsupported proto version: %s", version)
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
	case "", "latest", "2.1":
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

// DecodePositionWithID decodes a Position message and returns the sender node
// id along with the Position message. The id is zero when not present.
func DecodePositionWithID(data []byte, version string) (uint32, *latestpb.Position, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return 0, nil, err
	}
	switch version {
	case "", "latest", "2.1":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				from := pkt.GetFrom()
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_POSITION_APP {
						var pos latestpb.Position
						if err := proto.Unmarshal(dec.GetPayload(), &pos); err == nil {
							return from, &pos, nil
						}
					}
				}
			}
		}
		var pos latestpb.Position
		if err := proto.Unmarshal(data, &pos); err == nil && (pos.LatitudeI != nil || pos.LongitudeI != nil) {
			return 0, &pos, nil
		}
		return 0, nil, fmt.Errorf("not a Position message")
	default:
		return 0, nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}

// DecodeWaypointWithID decodes a Waypoint message and returns the sender node id
// along with the Waypoint. The id is zero when not present.
func DecodeWaypointWithID(data []byte, version string) (uint32, *latestpb.Waypoint, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return 0, nil, err
	}
	switch version {
	case "", "latest", "2.1":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				from := pkt.GetFrom()
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_WAYPOINT_APP {
						var wp latestpb.Waypoint
						if err := proto.Unmarshal(dec.GetPayload(), &wp); err == nil {
							return from, &wp, nil
						}
					}
				}
			}
		}
		var wp latestpb.Waypoint
		if err := proto.Unmarshal(data, &wp); err == nil && wp.GetId() != 0 {
			return 0, &wp, nil
		}
		return 0, nil, fmt.Errorf("not a Waypoint message")
	default:
		return 0, nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}

// DecodeNeighborInfoWithID decodes a NeighborInfo message and returns the sender
// node id along with the NeighborInfo. The id is zero when not present.
func DecodeNeighborInfoWithID(data []byte, version string) (uint32, *latestpb.NeighborInfo, error) {
	var err error
	data, err = stripFrame(data)
	if err != nil {
		return 0, nil, err
	}
	switch version {
	case "", "latest", "2.1":
		var fr latestpb.FromRadio
		if err := proto.Unmarshal(data, &fr); err == nil {
			if pkt := fr.GetPacket(); pkt != nil {
				from := pkt.GetFrom()
				if dec := pkt.GetDecoded(); dec != nil {
					if dec.GetPortnum() == latestpb.PortNum_NEIGHBORINFO_APP {
						var ni latestpb.NeighborInfo
						if err := proto.Unmarshal(dec.GetPayload(), &ni); err == nil {
							return from, &ni, nil
						}
					}
				}
			}
		}
		var ni latestpb.NeighborInfo
		if err := proto.Unmarshal(data, &ni); err == nil && ni.GetNodeId() != 0 {
			return 0, &ni, nil
		}
		return 0, nil, fmt.Errorf("not a NeighborInfo message")
	default:
		return 0, nil, fmt.Errorf("unsupported proto version: %s", version)
	}
}
