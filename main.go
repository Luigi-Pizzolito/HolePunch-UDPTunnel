package main

import (
	"flag"

	punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch"
	tui "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui"
)

const version = "1.0"

var (
	infoExchangeServer bool
)

func init() {
	// Parse command-line arguments
	flag.BoolVar(&infoExchangeServer, "server", false, "Run in info exchange server mode (run this on a publicly accesible IP)")
	flag.Parse()
}

func main() {
	// Setup TUI
	t := tui.Start(infoExchangeServer)
	t.Stage()
	// Setup logger.
	l := t.L
	l.Info("Hole Punch UDP Tunnel V"+version);
	

	s := punch.NewHPServer(l)
	go s.Hello()

	// Run TUI (blocking)
	t.RunApp()

	
}
