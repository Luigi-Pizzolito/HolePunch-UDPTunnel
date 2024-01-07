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
	connLogC chan string
	app *tview.Application
	serverMode bool
}

func Start(serverMode bool) *TUI {
	
	// Setup TUI
	app := tview.NewApplication().EnableMouse(true);

	// Setup log channel
	ch := make(chan string, 100);
	// Setup connection log channel
	cch := make(chan string, 20);

	return &TUI{L: nil, app: app, logC: ch, connLogC: cch, serverMode: serverMode}
}

func (t *TUI) Stage() {
	// Setup logger
	t.L = t.setupLogger();

	// Setup UI Flexbox layout
	// Populate UI elements in Flex
	flex := tview.NewFlex()
	t.app.SetRoot(flex, true).SetFocus(flex)

	// Setup UI elements
	if t.serverMode {
		// Holepunch server UI
		t.setupServerUI(flex)
	} else {
		// Holepunch client UI + UDP Tunnel UI
		t.setupClientUI(flex)
	}
	
}

func (t *TUI) setupSharedUI(flex *tview.Flex) {
	// Setup sub flex for stacked client and log view
	subflex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(subflex, 0, 4, false)

	// -- Client Info
	clientInfo := tview.NewTextView().
						SetDynamicColors(true).
						SetRegions(true).
						SetChangedFunc(func() {
							t.app.Draw()
						})
	clientInfo.SetBorder(true).SetTitle(" Client Info ")
	clientInfo.SetBackgroundColor(0)

	fmt.Fprintf(clientInfo, " [\"name\"]Name: Jesus[\"\"]\n")
	fmt.Fprintf(clientInfo, " [\"rAdr\"]Remote Address: udp://32.13.43.23:3213[\"\"]\n")
	if !t.serverMode {
		fmt.Fprintf(clientInfo, " [\"stat\"]Status: [green]Tunnel Active[-][\"\"]\n")
		fmt.Fprintf(clientInfo, " [\"lAdr\"]Local Address: 10.0.0.1[\"\"]\n")
		fmt.Fprintf(clientInfo, " [\"rPrt\"]Alowed Ports: 22, 80, 443[\"\"]\n")
		fmt.Fprintf(clientInfo, " [\"rPng\"]Ping: [green]32.4ms[-][\"\"]\n")
	}
	
	// add to subflex
	if t.serverMode {
		subflex.AddItem(clientInfo, 0, 2, false)
	} else {
		subflex.AddItem(clientInfo, 0, 4, false)
	}


	// -- Logs
	logs := tview.NewTextView().
					SetDynamicColors(true).
					SetRegions(true)
	logs.SetChangedFunc(func() {
		logs.ScrollToEnd()
		t.app.Draw()
	})
	logs.SetFocusFunc(func() {
		logs.ScrollToEnd()
	})
	logs.SetScrollable(true)
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

	// add to subflex
	subflex.AddItem(logs, 0, 6, false)
}

func (t *TUI) setupClientUI(flex *tview.Flex) {
	// -- Client List
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
	// add to flex
	flex.AddItem(clientList, 0, 2, true)

	// -- Shared UI elements (Client Info & Logs)
	t.setupSharedUI(flex)

	// -- Network Interfaces
	root := tview.NewTreeNode("Network Interfaces").
					SetColor(tcell.ColorRed).
					SetSelectable(false)
	activeTunnels := tview.NewTreeView().
							SetRoot(root).
							SetCurrentNode(nil)
	activeTunnels.SetBorder(true).SetTitle(" Active Tunnels ")
	activeTunnels.SetBackgroundColor(0)
	//----
	node := tview.NewTreeNode("tun0").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	root.AddChild(node)
	nnode := tview.NewTreeNode("10.0.0.1 ([blue]Jesus[-])").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	node.AddChild(nnode)
	//----
	node2 := tview.NewTreeNode("tun1").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	root.AddChild(node2)
	nnode2 := tview.NewTreeNode("10.0.0.2 ([blue]Lori[-])").
				// SetReference(filepath.Join(path, file.Name())).
				SetSelectable(false)
	node2.AddChild(nnode2)

	// add to flex 
	flex.AddItem(activeTunnels, 0, 2, false)

	// set app focus
	t.app.SetFocus(clientList)
}

func (t *TUI) setupServerUI(flex *tview.Flex) {
	// -- Client List
	clientList := tview.NewList().
		// ShowSecondaryText(false).
		AddItem("Jesus", "Idle", '1', nil).
		AddItem("Luigi", "[blue]Waiting for Celine", '2', nil).
		AddItem("Lori", "Idle", '3', func(){
			t.connLogC <- "[gray]06:19:58[-] Lori [blue]->[-] Luigi\n"
		})
	clientList.SetBorder(true).SetTitle(" Clients Online ")
	clientList.SetBackgroundColor(0)
	// add to flex
	flex.AddItem(clientList, 0, 2, true)

	// -- Shared UI elements (Client Info & Logs)
	t.setupSharedUI(flex)

	// -- Connection History
	connlogs := tview.NewTextView().
					SetDynamicColors(true).
					SetRegions(true)
	connlogs.SetChangedFunc(func() {
		connlogs.ScrollToEnd()
		t.app.Draw()
	})
	connlogs.SetFocusFunc(func() {
		connlogs.ScrollToEnd()
	})
	connlogs.SetScrollable(true)
	connlogs.SetBorder(true).SetTitle(" Connection History ")
	connlogs.SetBackgroundColor(0)
	
	// Go function to copy connlogs recieved in log channel to UI text box
	go func() {
		w := tview.ANSIWriter(connlogs)
		for {
			for str := range t.connLogC {
				if _, err := io.WriteString(w, str); err != nil {
					panic(err)
				}
			}
			
			time.Sleep(200 * time.Millisecond);
		}
	}()

	// add to flex
	flex.AddItem(connlogs, 0, 2, false)

	// set app focus
	t.app.SetFocus(clientList)
}

func (t *TUI) RunApp() {
	if err := t.app.Run(); err != nil {
		panic(err)
	}
}