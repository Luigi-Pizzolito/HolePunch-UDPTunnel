package main

import (
	"flag"
	"fmt"

	// punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch"
	tui "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui"
)

var (
	infoExchangeServer bool
)

func init() {
	// Parse command-line arguments
	flag.BoolVar(&infoExchangeServer, "s", false, "Run in info exchange server mode (run this on a publicly accesible IP)")
	flag.Parse()
}

func main() {
	fmt.Println("Hello world!")

	

	t := tui.Start()
	t.Stage()


	l := t.L
	l.Info("Still working?")
	

	// s := punch.NewHPServer(l)
	// go s.Hello()


	t.RunApp()

	fmt.Println("hi")


	
}
