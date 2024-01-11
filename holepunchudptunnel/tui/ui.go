package tui

import (
	"fmt"
	"io"
	"time"

	punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/natholepunch"
	tunnel "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/tunnelman"
	"strconv"
	"sort"

	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type TUI struct {
	// logging
	L *zap.Logger			// handle for main app logger
	Logfile string			// log file path
	logC chan string		// channel for redirecting main app log
	ConnLogC chan string	// channel for redirecting connection history log

	HPClientMap *map[string]punch.ClientData	// connection to HPServer/HPclient's client list
	UIupdate bool
	ClientUIStage2 bool

	// bind HPClient
	HPClient *punch.HPClient
	
	//todo: bind udptunnel tunnel info
	TunnelMan *tunnel.TunnelManager

	app *tview.Application	// application reference
	flex *tview.Flex
	// application elements reference
	// -- shared
	logs *tview.TextView
	clientInfo *tview.TextView
	clientList *tview.List
	selectedClient string
	// -- server only
	ConnLogs *tview.TextView

	serverMode bool			// server mode UI select
}

func Start(serverMode bool) *TUI {
	// Setup TUI
	app := tview.NewApplication().EnableMouse(true);

	// Setup log file
	log := "./log.log"
	// Setup log channel
	ch := make(chan string, 100);
	// Setup connection log channel
	cch := make(chan string, 20);

	return &TUI{
		L: 				nil,
		app: 			app,
		logC: 			ch,
		ConnLogC: 		cch,
		Logfile: 		log,
		serverMode: 	serverMode,
		UIupdate:		false,
		ClientUIStage2: false,
	}
}

func (t *TUI) Build() {
	// Setup logger
	t.L = t.setupLogger();

	// Setup UI Flexbox layout
	// Populate UI elements in Flex
	t.flex = tview.NewFlex()
	t.app.SetRoot(t.flex, true).SetFocus(t.flex)

	// Setup UI elements
	if t.serverMode {
		// Holepunch server UI
		t.setupServerUI(t.flex)
	} else {
		// Holepunch client UI + UDP Tunnel UI
		t.setupClientUI(t.flex)
	}

	// Automatically refresh UI element data using goroutine
	go t.refreshUIData()
	
}

func (t *TUI) setupSharedUI(flex *tview.Flex) {
	// Setup sub flex for stacked client and log view
	subflex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(subflex, 0, 4, false)

	// -- Client Info
	t.clientInfo = tview.NewTextView().
						SetDynamicColors(true).
						SetRegions(true).
						SetChangedFunc(func() {
							t.app.Draw()
						})
	t.clientInfo.SetBorder(true).SetTitle(" Client Info ")
	t.clientInfo.SetBackgroundColor(0)
	// add to subflex
	if t.serverMode {
		subflex.AddItem(t.clientInfo, 0, 2, false)
	} else {
		subflex.AddItem(t.clientInfo, 0, 4, false)
	}

	// -- Logs
	t.logs = tview.NewTextView().
					SetDynamicColors(true).
					SetRegions(true)
	t.logs.SetChangedFunc(func() {
		t.logs.ScrollToEnd()
		t.app.Draw()
	})
	t.logs.SetFocusFunc(func() {
		t.logs.ScrollToEnd()
	})
	t.logs.SetScrollable(true)
	t.logs.SetBorder(true).SetTitle(" Logs ")
	t.logs.SetBackgroundColor(0)
	// add to subflex
	subflex.AddItem(t.logs, 0, 6, false)
}

func (t *TUI) setupClientUI(flex *tview.Flex) {
	// -- Client List
	t.clientList = tview.NewList()
	t.clientList.SetBorder(true).SetTitle(" Clients Online ")
	t.clientList.SetBackgroundColor(0)
	// add to flex
	flex.AddItem(t.clientList, 0, 2, true)

	// -- Shared UI elements (Client Info & Logs)
	t.setupSharedUI(flex)

	// set app focus
	t.app.SetFocus(t.clientList)
}

func (t *TUI) setupServerUI(flex *tview.Flex) {
	// -- Client List
	t.clientList = tview.NewList()
	t.clientList.SetBorder(true).SetTitle(" Clients Queue ")
	t.clientList.SetBackgroundColor(0)
	// add to flex
	flex.AddItem(t.clientList, 0, 2, true)

	// -- Shared UI elements (Client Info & Logs)
	t.setupSharedUI(flex)

	// -- Connection History
	t.ConnLogs = tview.NewTextView().
					SetDynamicColors(true).
					SetRegions(true)
	t.ConnLogs.SetChangedFunc(func() {
		t.ConnLogs.ScrollToEnd()
		t.app.Draw()
	})
	t.ConnLogs.SetFocusFunc(func() {
		t.ConnLogs.ScrollToEnd()
	})
	t.ConnLogs.SetScrollable(true)
	t.ConnLogs.SetBorder(true).SetTitle(" Connection History ")
	t.ConnLogs.SetBackgroundColor(0)
	
	// Go function to copy t.ConnLogs recieved in log channel to UI text box
	go func() {
		w := tview.ANSIWriter(t.ConnLogs)
		for {
			for str := range t.ConnLogC {
				if _, err := io.WriteString(w, str); err != nil {
					panic(err)
				}
			}
			
			time.Sleep(200 * time.Millisecond);
		}
	}()
	// add to flex
	flex.AddItem(t.ConnLogs, 0, 2, false)

	// set app focus
	t.app.SetFocus(t.clientList)
}

// Client list linking functions
func (t *TUI) ConnectClientList(m *map[string]punch.ClientData) {
	t.HPClientMap = m;
}

func (t *TUI) getClientList() map[string]punch.ClientData {
	if t.HPClientMap == nil {
		return make(map[string]punch.ClientData)
	}
	return *t.HPClientMap
}

// HPClient linking functions
func (t *TUI) ConnectHPClient(c *punch.HPClient) {
	t.HPClient = c;
}

//todo: Tunnel Manager linking functions
func (t *TUI) ConnectTunMan(m *tunnel.TunnelManager) {
	t.TunnelMan = m;
}


// TUI data refresh function
func (t *TUI) refreshUIData() {
	// -- Logs
	// copy t.logs recieved in log channel to UI text box
	// requires separate goroutine to not block
	go func(){
		w := tview.ANSIWriter(t.logs)
		for str := range t.logC {
			if _, err := io.WriteString(w, str); err != nil {
				panic(err)
			}
		}
	}()

	// loop to always refresh data
	for {
		// Refresh data in UI elements for server & client UI specifically
		if t.serverMode {
			// Refresh server mode UI
			t.refreshServerUIData()
		} else {
			// Refresh client mode UI
			t.refreshClientUIData()
		}

		// Refresh data in shared mode UI elements
		t.refreshSharedUIData();

		// sleep to refresh at 10Hz
		time.Sleep(100 * time.Millisecond)
	}
}

func (t *TUI) refreshSharedUIData() {
	// -- Client Info
	// clear Client Info TextView
	t.clientInfo.SetText("")
	if t.clientList.GetItemCount() > 0 {
		// get currently selected client
		selectedCindex := t.clientList.GetCurrentItem();
		selectedCname, _ := t.clientList.GetItemText(selectedCindex);
		// get currently selected client info from HPServer/HPClient
		selectedC := t.getClientList()[selectedCname];
		// populate Client Info TextView
		if t.serverMode {
			// update client info
			fmt.Fprintf(t.clientInfo, " [\"name\"]Name: [green]%s[-][\"\"]\n", selectedC.LocalID)
			fmt.Fprintf(t.clientInfo, " [\"rAdr\"]Remote Address: [blue]udp://%s:%s[-][\"\"]\n", selectedC.LocalIP, selectedC.LocalPort)
		} else {
			// get currently selected client info from HPClient
			fmt.Fprintf(t.clientInfo, " [\"name\"]Name: [green]%s[-][\"\"]\n", selectedC.LocalID)
			fmt.Fprintf(t.clientInfo, " [\"rAdr\"]Remote Address: [blue]udp://%s:%s[-][\"\"]\n", selectedC.LocalIP, selectedC.LocalPort)
			
			// todo: populate actual data from TunnelManager
			tunClient := t.TunnelMan.TunClients[selectedCname]

			if tunClient.TunnelOn {
				fmt.Fprintf(t.clientInfo, " [\"stat\"]Status: [purple]Tunnel Active[-][\"\"]\n")
				fmt.Fprintf(t.clientInfo, " [\"lAdr\"]Local Address: [blue]10.0.0.1[-][\"\"]\n")
				fmt.Fprintf(t.clientInfo, " [\"rPrt\"]Alowed Ports: [blue]22, 80, 443[-][\"\"]\n")
			} else {
				fmt.Fprintf(t.clientInfo, " [\"stat\"]Status: [purple]Tunnel Inactive[-][\"\"]\n")
			}
			if tunClient.Ping != "" {
				fmt.Fprintf(t.clientInfo, " [\"rPng\"]Ping: [blue]"+tunClient.Ping+"[-][\"\"]\n")
			}

			fmt.Fprintf(t.clientInfo, " [red]%#v[-]\n", tunClient)
			
		}
	} else {
		// show that no clients are selected if there are no clients
		t.clientInfo.SetText("No clients selected.")
	}
}

// Function to refresh the TUI data
func (t *TUI) refreshServerUIData() {
	// Refresh data in HPServer mode UI elements

	if  t.UIupdate {
		// t.L.Info("Clientmap changed: "+strconv.FormatBool(mapChanged))

		// clear Clients Queue
		t.clientList.Clear()
		// populate Clients Queue
		sortedClients := sortClientsFromMap(t.getClientList())
		for i, client := range sortedClients {
			var status string
			if client.RemoteID == "" {
				status = "Idle"
			} else {
				status = "[blue]Waiting for "+client.RemoteID+"[-]"
			}
			t.clientList.AddItem(client.LocalID, status, []rune(strconv.Itoa(i+1))[0], nil)
		}
		t.UIupdate = false
		// t.L.Info("UI UPDATED!")
	}
}

func (t *TUI) refreshClientUIData() {
	// Refresh data in HPClient mode UI elements
	// -- Client List
	if t.UIupdate {
		// clear Clients Queue
		t.clientList.Clear()
		// populate Clients Queue
		sortedClients := sortClientsFromMap(t.getClientList())
		for i, client := range sortedClients {
			var status string
			clientID := client.LocalID
			//todo: Check tunnel status also
			if client.RemoteID == "" {
				status = "Idle"
			} else {
				//? is this the right place to put this?
				tunClient := t.TunnelMan.TunClients[clientID]
				if !tunClient.TunnelOn {
					status = "[blue]Waiting for "+client.RemoteID+"[-]"
				} else {
					status = "[purple]Tunnel Active[-]"
				}
			}
			t.clientList.AddItem(client.LocalID, status, []rune(strconv.Itoa(i+1))[0], func(){
				// this function is called when a client is selected
				t.requestClientConnect(clientID)
			})
		}

		t.UIupdate = false
		// t.L.Warn("UI UPDATED!")
	}

	// check if we are in client UI stage 2, if so change the UI layout
	if t.ClientUIStage2 {
		t.ClientUIStage2 = false

		t.flex.RemoveItem(t.clientList)
	}
}

// initiate client connection from HPClient
func (t *TUI) requestClientConnect(client string) {
	// This function is called when the TUI client selects another client to connect to

	// t.L.Info("Requesting connection to "+client)
	t.HPClient.InitiatePunch(client)

}

// Run TUI handler
func (t *TUI) RunApp() {
	if err := t.app.Run(); err != nil {
		panic(err)
	}
}

// utility functions
func sortClientsFromMap(m map[string]punch.ClientData) []punch.ClientData {
	// Collect keys from the map
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort keys alphabetically
	sort.Strings(keys)

	// Create a slice to store sorted structs
	sortedStructs := make([]punch.ClientData, len(keys))

	// Populate sortedStructs with struct values in sorted order
	for i, key := range keys {
		sortedStructs[i] = m[key]
	}

	return sortedStructs
}