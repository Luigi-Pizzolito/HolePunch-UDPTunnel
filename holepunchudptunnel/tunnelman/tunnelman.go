package tunnelman

import (
	"sort"
	"strings"
    "syscall"
    "golang.org/x/term"
	"io/ioutil"
	"os"
	"path/filepath"
	"go.uber.org/zap"
	"fmt"
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
func (m* TunnelManager) OpenTunnel(Self, SelfPort, Client string) {
	// create tunnel to Client
	// determine role
	role := m.determineRole(Self, Client)
	if role {
		m.l.Info("Determined to be tunnel server")
	} else {
		m.l.Info("Determined to be tunnel client")
	}
	// get sudo password //!add skip if user is already sudo
	m.stopMousePrinting()
	m.l.Warn("Elevating priviledges to open tunnel")
	passwd, err := m.getPasswd()
	if err != nil {
		m.l.Fatal("Failed to get sudo passwd")
		return
	}
	// m.l.Warn(passwd)
	m.resumeMousePrinting()
	// write udptunnel config file
	var config string
	filename := "udptunnel_config.json"
	if role {
		// Server config
		config = `
			{
				"TunnelAddress": "10.0.0.1",
				"NetworkAddress": ":`+SelfPort+`",
				"AllowedPorts": [22],
			}
		`
		m.l.Info("Tunnel server connected to :"+SelfPort)
	} else {
		// Client config
		config = `
			{
				"TunnelAddress": "10.0.0.2",
				"NetworkAddress": "`+m.TunClients[Client].EndPIP+`:`+m.TunClients[Client].EndPPort+`",
				"AllowedPorts": [22]
			}
		`
		m.l.Info("Tunnel client connected to "+m.TunClients[Client].EndPIP+`:`+m.TunClients[Client].EndPPort)
	}
	err = m.writeStringToFile(filename, config)
	if err != nil {
		m.l.Fatal("Error writing to udptunnel config:"+err.Error())
		return
	}
	m.l.Info("Wrote configuration file for UDP tunnel")

	//? Next: summon tunnel UDP with sudo
	m.l.Info("Starting Tunnel Daemon now")
	executeEmbeddedBinaryAsSudo(m.l, embeddedExecutable, "udptunnel", "udptunnel_config.json", passwd, "udptunnel_config.json")
	

}

func (m* TunnelManager) stopMousePrinting() {
    fmt.Print("\x1b[?1005l")
}

func (m* TunnelManager) resumeMousePrinting() {
    fmt.Print("\x1b[?1005h")
}

func (m* TunnelManager) determineRole(Self, Client string) bool {
	strSlice := []string{Self, Client}
	// Sort the slice of strings alphabetically
	sort.Strings(strSlice)
	// First string is the server
	if strSlice[0] == Self {
		// I am server
		return true
	} else {
		// I am client
		return false
	}
}

func (m* TunnelManager) getPasswd() (string, error) {
    // fmt.Print("Enter Password: ")
	m.l.Warn("Enter Sudo Password: ")
    bytePassword, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return "", err
    }
    password := string(bytePassword)
    return strings.TrimSpace(password), nil
}

func (m* TunnelManager) writeStringToFile(filename, content string) error {
	// Write the content to the file
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Get the directory of the executable
	exeDir := filepath.Dir(exePath)

	// Construct the file path relative to the executable directory
	filePath := filepath.Join(exeDir, filename)

	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
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