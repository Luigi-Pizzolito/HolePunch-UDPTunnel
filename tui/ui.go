package tui

import (
	"fmt"
	"io"
	// "os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type TUI struct {
	L *zap.Logger
	logC chan string
	app *tview.Application
}

func Start() *TUI {
	
	// Setup TUI
	app := tview.NewApplication().EnableMouse(true);

	// Setup log channel
	ch := make(chan string, 100);

	return &TUI{L: nil, app: app, logC: ch}
}

func (t *TUI) Stage() {
	// Setup logger
	t.L = t.setupLogger();

	// Setup UI elements
	clientList := tview.NewList().
		// ShowSecondaryText(false).
		AddItem("Jesus", "Tunnel Inactive", '1', nil).
		AddItem("Luigi", "[green]10.0.0.3", '2', nil).
		AddItem("Celine", "Tunnel Inactive", '3', nil).
		AddItem("Lori", "[red]10.0.0.2", '4', func() {
			t.L.Info("Lori Clicked")
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			t.app.Stop()
		})
	clientList.SetBorder(true).SetTitle(" Clients Online ")
	clientList.SetBackgroundColor(0)

	clientInfo := tview.NewTextView().
						SetDynamicColors(true).
						SetRegions(true).
						SetChangedFunc(func() {
							t.app.Draw()
						})
	clientInfo.SetBorder(true).SetTitle(" Client Info ")
	clientInfo.SetBackgroundColor(0)
	fmt.Fprintf(clientInfo, "[\"name\"]Name: Jesus[\"\"]\n")
	fmt.Fprintf(clientInfo, "[\"rAdr\"]Remote Address: udp://32.13.43.23:3213[\"\"]\n")
	fmt.Fprintf(clientInfo, "[\"stat\"]Status: [green]Tunnel Active[-][\"\"]\n")
	fmt.Fprintf(clientInfo, "[\"lAdr\"]Local Address: 10.0.0.1[\"\"]\n")
	fmt.Fprintf(clientInfo, "[\"rPrt\"]Alowed Ports: 22, 80, 443[\"\"]\n")
	fmt.Fprintf(clientInfo, "[\"rPng\"]Ping: [green]32.4ms[-][\"\"]\n")

	root := tview.NewTreeNode("Network Interfaces").
					SetColor(tcell.ColorRed).
					SetSelectable(false)
	activeTunnels := tview.NewTreeView().
							SetRoot(root).
							SetCurrentNode(root)
	activeTunnels.SetBorder(true).SetTitle(" Active Tunnels ")
	activeTunnels.SetBackgroundColor(0)
	//----
	node := tview.NewTreeNode("tun0")//.
				// SetReference(filepath.Join(path, file.Name())).
				// SetSelectable(true)
	root.AddChild(node)
	nnode := tview.NewTreeNode("10.0.0.1 ([blue]Jesus[-])").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	node.AddChild(nnode)
	//----
	node2 := tview.NewTreeNode("tun1")//.
				// SetReference(filepath.Join(path, file.Name())).
				// SetSelectable(true)
	root.AddChild(node2)
	nnode2 := tview.NewTreeNode("10.0.0.2 ([blue]Lori[-])").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	node2.AddChild(nnode2)

	logs := tview.NewTextView().
					SetDynamicColors(true).
					SetRegions(true).
					SetChangedFunc(func() {
						t.app.Draw()
					})
	logs.SetFocusFunc(func() {
		logs.ScrollToEnd()
	})
	logs.SetScrollable(false)
	logs.SetBorder(true).SetTitle(" Logs ")
	logs.SetBackgroundColor(0)
	
	// Go function to copy logs recieved in log channel to UI text box
	go func() {
		w := tview.ANSIWriter(logs)
		for {
			for str := range t.logC {
				if _, err := io.WriteString(w, str); err != nil {
					panic(err)
				}
			}
			
			time.Sleep(100 * time.Millisecond);
		}
	}()

	// Populate UI elements in Flex
	flex := tview.NewFlex().
		AddItem(clientList, 0, 2, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(clientInfo, 0, 4, false).
			AddItem(logs, 0, 6, false), 
			0, 4, false).
		AddItem(activeTunnels, 0, 2, false)
	t.app.SetRoot(flex, true).SetFocus(flex)
}

func (t *TUI) RunApp() {
	if err := t.app.Run(); err != nil {
		panic(err)
	}
}