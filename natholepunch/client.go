package natholepunch

import (
	"encoding/json"
	"time"
	"net"

	// tunnel "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/udptunnel"

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
		ClientList: 	make(map[string]ClientData),
	}
}

// Main logic
func (c *HPClient) Run() error {
	c.l.Info("Hole-Punch & UDP Tunnel Client mode")

	// Sync online clients list from server
	go func(){
		for {
			c.getClientList()
			//? maybe need to add a channel here to pause this when creating tunnel
			time.Sleep(1*time.Second)
		}
	}()

	//todo: accept commands here from TUI using a command channel



	return nil
}

// Teardown client class
func StopHPClient(l *zap.Logger, serverIP, serverPort, localID string) {
	// Stop UDP Tunnel
	// todo:

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
	_, err = conn.Write(requestJSON)
	if err != nil {
		l.Warn("Error disconnect request:"+err.Error())
		return
	}
}

// Get other clients listing
func (c *HPClient) getClientList() error {
	//todo: check if this is okat reusing the server UDP conn or not?
	
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Client List: Error resolving server address:"+err.Error())
		return err
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Client List: Error connecting to server:"+err.Error())
		return err
	}
	defer conn.Close()
	request := `{"local_id":"`+c.localID+`","remote_id":""}`	// client listing request
	requestJSON := []byte(request)
	_, err = conn.Write(requestJSON)
	if err != nil {
		c.l.Warn("Client List: Error disconnect request:"+err.Error())
		return err
	}
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		c.l.Warn("Client List: Error receiving data:"+err.Error())
		return err
	}

	out := make(map[string]ClientData)
	err = json.Unmarshal(buffer[:n], &out)
	if err != nil {
		c.l.Warn("Client List: Error parsing data:"+err.Error())
		return err
	}

	// c.ClientList = make(map[string]ClientData, len(out))
	// for k, v := range out {
	// 	c.ClientList[k] = v
	// }
	c.ClientList = out;
	return nil
}


// Client punching functions to start the connection

// Contact info exchange server, perform hole punch and test connection with ping
func (c *HPClient) InitiatePunch(client string) {
	c.l.Info("Request for hole punch to "+client)

	// set our clients remoteID
	c.RemoteID = client;
	// check if we are ready to hole punch (other client also wants)
	clientData := c.ClientList[client];
	if clientData.RemoteID == c.localID {
		c.l.Info("Remote client also wants to connect")
		c.l.Info("Initiating hole punch to "+client)
		c.pingNpunch()
	} else {
		c.l.Warn("Remote client does not want to connect") 
		if clientData.RemoteID == "" {
			c.l.Info("Remote client is Idle")
		} else {
			c.l.Info("Remote client is waiting for "+clientData.RemoteID)
		}
		c.l.Info("Please ask remote client to connect to you too")
		return
	}
}

func (c *HPClient) pingNpunch() {

}

// createP2PConnection sends local and remote ID to the server and waits for the response
func (c *HPClient) createP2PConnection() {

}

// REconnect allows to manually send reconnect signal, returns new remote addr and port
func (c *HPClient) REconnect() (*net.UDPAddr, error) {

}



//! both clients need to request connection for hole punch to work?? check in network test

//!on disconnect client must send connection request with null-byte as remoteID