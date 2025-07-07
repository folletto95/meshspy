//go:build gui

package gui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	mqttpkg "meshspy/client"
	"meshspy/config"
	"meshspy/nodemap"
	latestpb "meshspy/proto/latest/meshtastic"
	"meshspy/serial"
	"meshspy/storage"
)

// Run launches a desktop UI that can connect to a serial port and display
// messages and nodes similarly to the web interface.
func Run(cfg config.Config, nodeStore *storage.NodeStore) {
	a := app.New()
	w := a.NewWindow("MeshSpy GUI")

	serialEntry := widget.NewEntry()
	serialEntry.SetText(cfg.SerialPort)
	connectBtn := widget.NewButton("Connect", nil)
	disconnectBtn := widget.NewButton("Disconnect", nil)

	messages := widget.NewMultiLineEntry()
	messages.Wrapping = fyne.TextWrapWord
	messages.Disable()

	sendEntry := widget.NewEntry()
	sendBtn := widget.NewButton("Send", nil)

	var (
		mgr *serial.Manager
		mu  sync.Mutex
		nm  = nodemap.New()
	)

	// display nodes from the DB in a scrolling list
	var nodeData []*mqttpkg.NodeInfo
	updateNodes := func() {
		nodes, err := nodeStore.List()
		if err != nil {
			return
		}
		nodeData = nodes
		list.Refresh()
	}

	list := widget.NewList(
		func() int { return len(nodeData) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, o fyne.CanvasObject) {
			if i < len(nodeData) {
				n := nodeData[i]
				name := n.LongName
				if name == "" {
					name = n.ID
				}
				o.(*widget.Label).SetText(
					fmt.Sprintf("%s [%s] id:%s battery:%d%%", name, n.ShortName, n.ID, n.BatteryLevel))
			}
		})

	updateNodes()

	connect := func() {
		port := serialEntry.Text
		m, err := serial.OpenManager(port, cfg.BaudRate)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		mu.Lock()
		mgr = m
		mu.Unlock()

		go mgr.ReadLoop(cfg.Debug, "", nm,
			func(ni *latestpb.NodeInfo) {
				info := mqttpkg.NodeInfoFromProto(ni)
				if info != nil {
					if err := nodeStore.Upsert(info); err == nil {
						a.QueueUpdate(updateNodes)
					}
				}
			},
			nil,
			func(txt string) {
				a.QueueUpdate(func() {
					messages.SetText(messages.Text + txt + "\n")
				})
			},
			func(payload string) {})
	}

	connectBtn.OnTapped = connect
	disconnectBtn.OnTapped = func() {
		mu.Lock()
		if mgr != nil {
			mgr.Close()
			mgr = nil
		}
		mu.Unlock()
	}

	sendBtn.OnTapped = func() {
		mu.Lock()
		defer mu.Unlock()
		if mgr == nil {
			dialog.ShowInformation("Not connected", "Serial port not open", w)
			return
		}
		if err := mgr.SendTextMessage(sendEntry.Text); err != nil {
			dialog.ShowError(err, w)
		} else {
			sendEntry.SetText("")
		}
	}

	top := container.NewHBox(widget.NewLabel("Serial:"), serialEntry, connectBtn, disconnectBtn)
	sendBox := container.NewHBox(sendEntry, sendBtn)
	w.SetContent(container.NewVBox(
		top,
		widget.NewLabel("Messages"),
		container.NewScroll(messages),
		widget.NewLabel("Send"),
		sendBox,
		widget.NewLabel("Nodes"),
		container.NewVScroll(list),
	))
	w.Resize(fyne.NewSize(600, 400))
	w.ShowAndRun()
}
