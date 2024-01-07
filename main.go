package main

import (
	"flag"

	punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch"
	tui "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui"
)

const version = "1.0"

var (
	// Server mode switch
	infoExchangeServer bool
	// Server mode variables
	serverAddr string
	serverPort string
)

func init() {
	// Parse command-line arguments
	flag.BoolVar(&infoExchangeServer, "server", false, "Run in info exchange server mode (run this on a publicly accesible IP)")
	flag.StringVar(&serverPort, "server-port", "10001", "Info exchange server port to listen on.")

	flag.StringVar(&serverAddr, "info-server", "127.0.0.1", "Info exchange server IP address to bind to.")
	flag.Parse()
}

func main() {
	// Setup TUI
	t := tui.Start(infoExchangeServer)
	t.Stage()
	// Setup logger.
	l := t.L
	l.Info("Hole Punch UDP Tunnel V"+version);
	
	if infoExchangeServer {
		// Info Exchange Server mode
		// Start server
		s := punch.NewHPServer(l)
		go s.Serve(serverAddr,  serverPort)
	} else {
		// Holepunch + UDP Tunnel client mode
	}

	// Run TUI (blocking)
	t.RunApp()
}
