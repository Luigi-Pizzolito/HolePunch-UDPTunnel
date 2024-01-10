package tunnelman

// ClientTunnelData has all data about client's local tunnel status
type ClientTunnelData struct {
	TunnelOn	bool
	TunnelAddr	string
	TunnelPorts	[]int
	Ping		string
	// for udp tunnel
	EndPIP		string
	EndPPort	string
	EndPAPorts	[]int
}