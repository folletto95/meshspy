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

func ExampleMap_UpdateFromProto_ignoreEmpty() {
	nm := New()
	nm.UpdateFromProto(&latestpb.NodeInfo{Num: 0x0, User: &latestpb.User{}})
	fmt.Println(nm.Resolve("0x0"))
	// Output: 0x0
}
