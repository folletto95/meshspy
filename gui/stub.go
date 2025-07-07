//go:build !gui

package gui

import (
	"meshspy/config"
	"meshspy/storage"
)

// Run is a no-op when built without the gui tag.
func Run(_ config.Config, _ *storage.NodeStore) {}
