module github.com/Luigi-Pizzolito/HolePunch-UDPTunnel

go 1.21.5

replace github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch => ./natholepunch

replace github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui => ./tui

replace github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/udptunnel => ./tunnel

require (
	github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/natholepunch v0.0.0-00010101000000-000000000000
	github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui v0.0.0-00010101000000-000000000000
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell/v2 v2.6.1-0.20231203215052-2917c3801e73 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/tview v0.0.0-20240101144852-b3bd1aa5e9f2 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/term v0.9.0 // indirect
	golang.org/x/text v0.12.0 // indirect
)
