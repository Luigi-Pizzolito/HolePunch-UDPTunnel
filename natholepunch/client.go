package natholepunch

import (
	// "encoding/json"
	"time"
	"net"
	// "fmt"

	"go.uber.org/zap"
)

// Define the Hole Punch client struct
type HPClient struct {
	// Channels for signaling status
	connectStatus 	chan error
	stopChan	  	chan struct{}
	// Connections and response
	Conn 			*net.UDPConn
	R				*ConnectResponse
	timeout			time.Duration
	// Client info
	serverAddr		string
	serverPort		string
	localID			string
	RemoteID		string
	// Client map from HPServer
	ClientList		map[string]ClientData
	//todo: Tunnel info & udptunnel command channel
	//TunnelInfo		map...
	//TunnelControl		chan ?
	// Logger
	l				*zap.Logger
}

// Initialises a pointer to a new HPClient struct
func NewHPClient(l *zap.Logger, timeout time.Duration, serverAddr, serverPort, localID, remoteID string/*, bind UI tunnel listing here later*/) *HPClient {
	return &HPClient{
		connectStatus:	make(chan error),
		stopChan:		make(chan struct{}),
		timeout:		timeout,
		serverAddr:		serverAddr,
		serverPort:		serverPort,
		localID:		localID,
		RemoteID:		remoteID,
		l:				l,
	}
}

// Main logic
func (c *HPClient) Run() error {
	c.l.Info("Hole-Punch & UDP Tunnel Client mode")

	// Sync online clients list from server
	// go getClientList

	// 



	return nil
}

// Teardown client class
func StopHPClient(l *zap.Logger, serverIP, serverPort, localID string) {
	// Stop UDP Tunnel

	// Send disconnect packet to server
	//todo: add error handling print msgs here
	serverAddr, err := net.ResolveUDPAddr("udp", serverIP+":"+serverPort)
	if err != nil {
		l.Warn("Error resolving server address:"+err.Error())
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		l.Warn("Error connecting to server:"+err.Error())
		return
	}
	defer conn.Close()

	request := `{"local_id":"`+localID+`","remote_id":"\u0000"}`	// disconnect request
	requestJSON := []byte(request)

	// _, err = conn.WriteToUDP(requestJSON, serverAddr)
	_, err = conn.Write(requestJSON)
	if err != nil {
		l.Warn("Error disconnect request:"+err.Error())
		return
	}
}
/*
// Get other clients listing
func (c *HPClient) getClientList() map[string]ClientData {

}

// Contact info exchange server, perform hole punch and test connection with ping
func (c *HPClient) punchNping() {

}

// createP2PConnection sends local and remote ID to the server and waits for the response
func (c *Client) createP2PConnection() {

}

// REconnect allows to manually send reconnect signal, returns new remote addr and port
func (c *Client) REconnect() (*net.UDPAddr, error) {

}
*/


//! both clients need to request connection for hole punch to work?? check in network test

//!on disconnect client must send connection request with null-byte as remoteID