//go:build gui

package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"meshspy/storage"
)

// Run launches a simple GUI showing the list of nodes stored in nodeStore.
func Run(nodeStore *storage.NodeStore) {
	nodes, err := nodeStore.List()
	if err != nil {
		fmt.Printf("failed to load nodes: %v\n", err)
		return
	}

	a := app.New()
	w := a.NewWindow("MeshSpy Nodes")

	var items []fyne.CanvasObject
	for _, n := range nodes {
		items = append(items, widget.NewLabel(fmt.Sprintf("%s - %s", n.ID, n.LongName)))
	}

	w.SetContent(container.NewVBox(items...))
	w.ShowAndRun()
}
