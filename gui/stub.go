//go:build !gui

package gui

import "meshspy/storage"

// Run is a no-op when built without the gui tag.
func Run(_ *storage.NodeStore) {}
