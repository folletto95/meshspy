package nodemap

import (
	"fmt"
	"sort"
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

// Node represents a node entry along with its identifier.
type Node struct {
	ID string
	Entry
}

func New() *Map {
	return &Map{nodes: make(map[string]Entry)}
}

func (m *Map) Update(num uint32, long, short string) {
	id := fmt.Sprintf("0x%x", num)
	m.mu.Lock()
	e := m.nodes[id]
	if long != "" {
		e.Long = long
	}
	if short != "" {
		e.Short = short
	}
	m.nodes[id] = e
	m.mu.Unlock()
}

func (m *Map) UpdateFromProto(ni *latestpb.NodeInfo) {
	if ni == nil || ni.User == nil {
		return
	}
	if ni.GetNum() == 0 && ni.User.GetLongName() == "" && ni.User.GetShortName() == "" {
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

// ResolveLong returns the long name for the given node id when available.
// If the long name is empty, it falls back to the short name and finally to
// the id itself when no name is known.
func (m *Map) ResolveLong(id string) string {
	m.mu.RLock()
	e, ok := m.nodes[id]
	m.mu.RUnlock()
	if ok {
		if e.Long != "" {
			return e.Long
		}
		if e.Short != "" {
			return e.Short
		}
	}
	return id
}

// List returns a snapshot of all known nodes sorted by id.
func (m *Map) List() []Node {
	m.mu.RLock()
	nodes := make([]Node, 0, len(m.nodes))
	for id, e := range m.nodes {
		nodes = append(nodes, Node{ID: id, Entry: e})
	}
	m.mu.RUnlock()
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes
}
