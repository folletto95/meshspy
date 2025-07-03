package nodemap

import (
	"fmt"

	latestpb "meshspy/proto/latest/meshtastic"
)

func ExampleMap_UpdateFromProto() {
	nm := New()
	nm.UpdateFromProto(&latestpb.NodeInfo{
		Num:  0x1234,
		User: &latestpb.User{LongName: "Alice", ShortName: "A"},
	})
	fmt.Println(nm.Resolve("0x1234"))
	// Output: Alice
}
