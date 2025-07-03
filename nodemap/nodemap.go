package nodemap

import (
	"fmt"
	"sync"

	latestpb "meshspy/proto/latest/meshtastic"
)

type Entry struct {
	Long  string
	Short string
}

type Map struct {
	mu    sync.RWMutex
	nodes map[string]Entry
}

func New() *Map {
	return &Map{nodes: make(map[string]Entry)}
}

func (m *Map) Update(num uint32, long, short string) {
	id := fmt.Sprintf("0x%x", num)
	m.mu.Lock()
	m.nodes[id] = Entry{Long: long, Short: short}
	m.mu.Unlock()
}

func (m *Map) UpdateFromProto(ni *latestpb.NodeInfo) {
	if ni == nil || ni.User == nil {
		return
	}
	m.Update(ni.GetNum(), ni.User.GetLongName(), ni.User.GetShortName())
}

func (m *Map) Resolve(id string) string {
	m.mu.RLock()
	e, ok := m.nodes[id]
	m.mu.RUnlock()
	if !ok {
		return id
	}
	if e.Long != "" {
		return e.Long
	}
	if e.Short != "" {
		return e.Short
	}
	return id
}

// ResolveLong returns the long name for the given node id if known, otherwise
// it falls back to the id itself.
func (m *Map) ResolveLong(id string) string {
	m.mu.RLock()
	e, ok := m.nodes[id]
	m.mu.RUnlock()
	if ok && e.Long != "" {
		return e.Long
	}
	return id
}
