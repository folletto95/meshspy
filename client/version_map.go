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

	parts := strings.SplitN(fw, ".", 3)
	if len(parts) >= 2 {
		major, minor := parts[0], parts[1]
		// Firmware 2.1.x uses the older protobuf schema
		if major == "2" && minor == "1" {
			return "2.1"
		}
	}

	return "latest"
}
