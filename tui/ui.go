package tui

import (
	"fmt"
	"io"
	"os"
	// "time"

	"https://github.com/fsnotify/fsnotify"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type TUI struct {
	L *zap.Logger
	// l *chan string
	Logstream *TBuffer
	app *tview.Application
}

func Start() *TUI {
	// Start logger
	logstream := &TBuffer{}
	
	// Setup TUI
	app := tview.NewApplication().EnableMouse(true)

	
	memfile := NewMemFile([]byte("log"))

	return &TUI{L: nil, Logstream: logstream, memfile: memfile, app: app}
}

func (t *TUI) Stage() {
	t.L = t.setupLogger();

	
	
	t.L.Info("Hellow");

	
	clientList := tview.NewList().
		// ShowSecondaryText(false).
		AddItem("Jesus", "Tunnel Inactive", '1', nil).
		AddItem("Luigi", "[green]10.0.0.3", '2', nil).
		AddItem("Celine", "Tunnel Inactive", '3', nil).
		AddItem("Lori", "[red]10.0.0.2", '4', func() {
			t.L.Info("Lori Clicked")
			// t.Logstream.WriteString("test4\n")
			// t.app.Draw()
			// t.Logs.SetText("aaa")
			// t.Logstream.WriteString("asdf");
			// go t.memfile.Write([]byte("asdfff\n"));
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			t.app.Stop()
		})
	clientList.SetBorder(true).SetTitle(" Clients Online ")

	clientInfo := tview.NewTextView().
						SetDynamicColors(true).
						SetRegions(true).
						SetChangedFunc(func() {
							t.app.Draw()
						})
	clientInfo.SetBorder(true).SetTitle(" Client Info ")
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
		fmt.Println("test3")
		logs.ScrollToEnd()
	})
	logs.SetScrollable(false)
	logs.SetBorder(true).SetTitle(" Logs ")


	

	go func() {
		w := tview.ANSIWriter(logs)
		if _, err := io.Copy(w, os.Stdin); err != nil {
			panic(err)
		}
		fmt.Printf("ANSI Writer exited.\n")
	}()
	fmt.Fprintf(logs, "logs here\nand here\nand here too\nand here as well\n")
	go func(){
	// 	w := logs.BatchWriter()
	// 	defer w.Close()
	// 	// w.Clear()
	// 	for {
	// // 		fmt.Fprintln(w, "To sit in solemn silence")
	// // 	// for {
	// // 	// 	// msg := <-logChn
    // // 	// 	fmt.Println(t.Logstream.String())
	// // 	// 	// fmt.Fprintf(t.Logstream, "%s", msg)
	// // 	// 	// t.Logstream.WriteString("asdf");
	// 		time.Sleep(1*time.Second)
	// 	}
	}()


	flex := tview.NewFlex().
		AddItem(clientList, 0, 2, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(clientInfo, 0, 6, false).
			AddItem(logs, 0, 4, false), 
			0, 4, false).
		AddItem(activeTunnels, 0, 2, false)
	t.app.SetRoot(flex, true).SetFocus(flex)
}

func (t *TUI) RunApp() {
	if err := t.app.Run(); err != nil {
		panic(err)
	}
}