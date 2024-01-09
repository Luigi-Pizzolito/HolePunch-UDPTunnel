package udptunnel

import (
	"go.uber.org/zap"
)

// Define the Tunnel Manager Struct
type TunnelManager struct {
	// Logs
	l				*zap.Logger
	// keep list of clients and their tunnel infos
	TunClients		map[string]ClientTunnelData
}

// Initiallises a pointer to a new TunnelManager struct
func NewTunnelManager(l *zap.Logger) *TunnelManager {
	return &TunnelManager{
		l:				l,
		TunClients:		make(map[string]ClientTunnelData),
	}
}

// Main logic
func (m* TunnelManager) Run() error {
	return nil
}

// Allow client info to be populated by HPClient
func (m* TunnelManager) AddClient(Client, TunnelAddr string, TunnelPorts []int, EndPIP, EndPPort string, EndPAPorts []int, Ping string) {
	m.TunClients[Client] = ClientTunnelData{
		TunnelOn:		false,
		TunnelAddr:		TunnelAddr,
		TunnelPorts:	TunnelPorts,
		Ping:			Ping,
		EndPIP:			EndPIP,
		EndPPort:		EndPPort,
		EndPAPorts:		EndPAPorts,
	}
}

// Commands
func (m* TunnelManager) OpenTunnel(Client string) {
	// create tunnel to Client
}

func (m* TunnelManager) CloseTunnel(Client string) {
	// close tunnel to Client
}

//? main.go creates a NewTunnelManager
//? and passes it to TUI and HPClient by pointer
//? then TUI can:
//? 	access m.TunClients to display info
//? and HPClient can:
//? 	use use m.AddClient(Client, TunnelAddr, TunnelPorts, EndPIP, EndPPort, EndPAPorts, Ping)
//? 		to populate all client info from HPClient into m.TunClients
//?		use.Command channel to forward connect and disconnect commands
//?			(sent to HPClient from TUI) to TunnelManager