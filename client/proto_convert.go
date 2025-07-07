package mqtt

import (
	"fmt"

	latestpb "meshspy/proto/latest/meshtastic"
)

// NodeInfoFromProto converts a protobuf NodeInfo message to the internal NodeInfo type.
func NodeInfoFromProto(ni *latestpb.NodeInfo) *NodeInfo {
	if ni == nil || ni.GetUser() == nil {
		return nil
	}
	u := ni.GetUser()
	info := &NodeInfo{
		ID:        fmt.Sprintf("0x%x", ni.GetNum()),
		Num:       ni.GetNum(),
		LongName:  u.GetLongName(),
		ShortName: u.GetShortName(),
		MacAddr:   fmt.Sprintf("%x", u.GetMacaddr()),
		HwModel:   u.GetHwModel().String(),
		Role:      u.GetRole().String(),
	}
	if pos := ni.GetPosition(); pos != nil {
		info.Latitude = float64(pos.GetLatitudeI()) / 1e7
		info.Longitude = float64(pos.GetLongitudeI()) / 1e7
		info.Altitude = int(pos.GetAltitude())
		info.LocationTime = int64(pos.GetTime())
		info.LocationSource = pos.GetLocationSource().String()
	}
	if dm := ni.GetDeviceMetrics(); dm != nil {
		info.BatteryLevel = int(dm.GetBatteryLevel())
		info.Voltage = float64(dm.GetVoltage())
		info.ChannelUtil = float64(dm.GetChannelUtilization())
		info.AirUtilTx = float64(dm.GetAirUtilTx())
		info.UptimeSeconds = int(dm.GetUptimeSeconds())
	}
	info.Snr = float64(ni.GetSnr())
	info.LastHeard = int64(ni.GetLastHeard())
	info.Channel = int(ni.GetChannel())
	info.ViaMqtt = ni.GetViaMqtt()
	info.HopsAway = int(ni.GetHopsAway())
	info.IsFavorite = ni.GetIsFavorite()
	info.IsIgnored = ni.GetIsIgnored()
	info.IsKeyManuallyVerified = ni.GetIsKeyManuallyVerified()
	return info
}

// NodeInfoFromMyInfo converts a MyNodeInfo message to the internal NodeInfo type.
func NodeInfoFromMyInfo(mi *latestpb.MyNodeInfo) *NodeInfo {
	if mi == nil {
		return nil
	}
	return &NodeInfo{
		ID:  fmt.Sprintf("0x%x", mi.GetMyNodeNum()),
		Num: mi.GetMyNodeNum(),
	}
}
