package main

import (
	"flag"
	"go.uber.org/zap"
	"fmt"
	"time"
	"os/user"

	punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/natholepunch"
	tunnel "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/tunnelman"
	tui "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/tui"
)

const version = "0.2"

var (
	// Server mode switch
	infoExchangeServer 	bool
	// Server mode variables
	serverPort 			string
	// CLient mode variables
	infoAddr 			string
	localID 			string
	remoteID			string
	timeout				time.Duration
)

// function to initialise before main
func init() {
	// Parse command-line arguments
	// -- Server Mode flags
	flag.BoolVar(&infoExchangeServer, "server", false, "Run in info exchange server mode (run this on a publicly accesible IP)")
	flag.StringVar(&serverPort, "p", "10001", "Info exchange server port to listen on or connect to")
	// -- Client Mode flags
	flag.StringVar(&infoAddr, "a", "127.0.0.1", "Info exchange server IP address to connect to")
	// Get current OS username as default Local side ID
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&localID, "l", currentUser.Username, "Specify Local side ID")
	flag.StringVar(&remoteID, "r", "", "(optional) Specify remote side ID to tunnel to")
	flag.DurationVar(&timeout, "t", time.Second*2, "Specify reconnection timeout")
	flag.Parse()
}

// function to cleanup after main
func teardown(l *zap.Logger, serverMode bool) {
	l.Warn("Tearing down application!")
	fmt.Println("Tearing down application!")

	// stop terminal mouse printing
	tunnel.StopMousePrinting()

	if !serverMode {
		// Client mode termination
		//Send a request to server to inform client going offline
		punch.StopHPClient(l, infoAddr, serverPort, localID)
	}
}

func main() {
	//todo: add graceful exit

	// Setup TUI
	t := tui.Start(infoExchangeServer)
	t.Build()

	// Setup logger.
	l := t.L
	fmt.Println("Hole Punch UDP Tunnel V"+version);
	l.Info("Hole Punch UDP Tunnel V"+version);
	fmt.Println("Log file at "+t.Logfile)

	// Setup app teardown
	defer teardown(l, infoExchangeServer)
	
	// Start main logic
	if infoExchangeServer {

		// Info Exchange Server mode
		// Start server
		s := punch.NewHPServer(l, t.ConnLogC, &t.UIupdate)
		t.ConnectClientList(&s.ClientList)
		go func(){
			if err := s.Serve(serverPort); err != nil {
				l.Fatal(err.Error())
			}
		}()

	} else {

		// Holepunch + UDP Tunnel client mode
		m := tunnel.NewTunnelManager(l)
		c := punch.NewHPClient(l, timeout, infoAddr, serverPort, localID, remoteID, &t.UIupdate, &t.ClientUIStage2, m)
		t.ConnectClientList(&c.ClientList)
		t.ConnectHPClient(c)
		t.ConnectTunMan(m)
		//? add t.Run() here
		go func(){
			if err := c.Run(); err != nil {
				l.Fatal(err.Error())
			}
		}()

	}

	// Run TUI (blocking)
	t.RunApp()
}
