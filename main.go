package main

import (
	"flag"
	"fmt"

	"time"

	punch "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch"
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


	go func() {
		l.Info("Hello from goroutine!")
		i := 0
		for {
			// Display integer.
			i++
			// l.Info(strconv.Itoa(i))
			// logChn <- strconv.Itoa(i)
			// t.Logs.SetText(strconv.Itoa(i))
			time.Sleep(1 * time.Second)
    	}
	}()
	

	s := punch.NewHPServer(l)
	go s.Hello()

	// t.Logs.SetText("aaa")

	t.RunApp()

	fmt.Println("hi")


	
}
