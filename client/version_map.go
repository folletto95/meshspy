package mqtt

import "strings"

// ProtoVersionForFirmware returns the protobuf schema version to use
// when communicating with a device running the given firmware version.
// Currently it returns "latest" for all versions but allows future
// mappings to be added.
func ProtoVersionForFirmware(fw string) string {
	if fw == "" {
		return "latest"
	}
	// example: 2.1.x might require old proto, here we just default
	if strings.HasPrefix(fw, "2.") {
		return "latest"
	}
	return "latest"
}
