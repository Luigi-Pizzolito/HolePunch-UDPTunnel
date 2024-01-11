package natholepunch

import (
	"encoding/json"
	"time"
	"net"
	"fmt"
	"errors"
	"bytes"
	"strconv"

	tunnel "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/holepunchudptunnel/tunnelman"

	"go.uber.org/zap"
)

// Define the Hole Punch client struct
type HPClient struct {
	// Channels for signaling status
	connectStatus 	chan error
	stopChan	  	chan struct{}
	pauseClientFetch bool
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
	// UI update hook
	UIupdate		*bool
	ClientUIStage2	*bool
	UIupdateCListB	[]byte
	// Tunnel man interface
	TunnelMan			*tunnel.TunnelManager
	// Logger
	l				*zap.Logger
}

// Initialises a pointer to a new HPClient struct
func NewHPClient(l *zap.Logger, timeout time.Duration, serverAddr, serverPort, localID, remoteID string, UIupdate, ClientUIStage2 *bool, TunnelMan *tunnel.TunnelManager) *HPClient {
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
		UIupdate:		UIupdate,
		ClientUIStage2: ClientUIStage2,
		UIupdateCListB: make([]byte, 1024),
		pauseClientFetch: false,

		TunnelMan:			TunnelMan,
	}
}

// Main logic
func (c *HPClient) Run() error {
	c.l.Info("Hole-Punch & UDP Tunnel Client mode")

	// Sync online clients list from server
	go func(){
		for {
			// check flag here to pause updating client list
			if !c.pauseClientFetch {
				c.getClientList()
			}
			time.Sleep(1*time.Second) // update every second
		}
	}()

	return nil
}

// Teardown client class
func StopHPClient(l *zap.Logger, serverIP, serverPort, localID string) {
	/
	// Send disconnect packet to server
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
	
	// Resolve server adress
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Client List: Error resolving server address:"+err.Error())
		return err
	}

	// Start connection to server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Client List: Error connecting to server:"+err.Error())
		return err
	}
	defer conn.Close()

	// Create request JSON and send request
	request := `{"local_id":"`+c.localID+`","remote_id":""}`
	requestJSON := []byte(request)
	_, err = conn.Write(requestJSON)
	if err != nil {
		c.l.Warn("Client List: Error disconnect request:"+err.Error())
		return err
	}

	// Read server response
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		c.l.Warn("Client List: Error receiving data:"+err.Error())
		return err
	}

	// Check if the client list has changed, if so update UI
	if !bytes.Equal(c.UIupdateCListB, buffer) {
		// if the response is different, trigger an UI update
		*c.UIupdate = true;
		// copy the new response to the change buffer to detect new change
		copy(c.UIupdateCListB, buffer)
	}

	// Parse response JSON
	out := make(map[string]ClientData)
	err = json.Unmarshal(buffer[:n], &out)
	if err != nil {
		c.l.Warn("Client List: Error parsing data:"+err.Error())
		return err
	}
	c.ClientList = out;

	return nil
}

func (c *HPClient) updateServerRemoteID() error {
	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Client List: Error resolving server address:"+err.Error())
		return err
	}

	// Start UDP connection
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Client List: Error connecting to server:"+err.Error())
		return err
	}
	defer conn.Close()

	// Create request JSON
	request := `{"local_id":"`+c.localID+`","remote_id":"`+c.RemoteID+`"}`	// client listing request when idle as c.RemoteID=="", otherwise update with waiting
	requestJSON := []byte(request)
	_, err = conn.Write(requestJSON)
	if err != nil {
		c.l.Warn("Client List: Error disconnect request:"+err.Error())
		return err
	}

	return nil
}


// Client punching functions to start the connection

// Contact info exchange server, perform hole punch and test connection with ping
func (c *HPClient) InitiatePunch(client string) {

	// set flag to change client UI to stage 2
	*c.ClientUIStage2 = true

	c.l.Info("Request for hole punch to "+client)

	// set our clients remoteID
	c.pauseClientFetch = true;
	c.RemoteID = client;
	// update server with remoteID for connection request
	c.updateServerRemoteID()
	*c.UIupdate = true; // update UI
	

	// check if we are ready to hole punch (other client also wants to connect)
	clientData := c.ClientList[client]
	if clientData.RemoteID == c.localID {
		c.l.Info("Remote client also wants to connect")
		c.l.Info("Initiating hole punch to "+client)
		// Initiate punch by calling pingNpunch
		if err := c.pingNpunch(); err != nil {
			c.l.Error(err.Error())
			return
		}
		c.l.Info("Hole punching to "+client)

	} else {
		// Other client does not want to connect
		c.l.Warn("Remote client does not want to connect")  
		// Print other client status
		if clientData.RemoteID == "" {
			c.l.Info("Remote client is Idle")
		} else {
			c.l.Info("Remote client is waiting for "+clientData.RemoteID)
		}
		c.l.Info("Please ask remote client to connect to you too")

		// Here we pause the client fetch
		c.pauseClientFetch = true;
		// Then we wait in a loop for the other client to change status to accept our connection
		go func(c *HPClient) {
			c.l.Info("Waiting in loop for "+c.RemoteID)
			for c.ClientList[c.RemoteID].RemoteID != c.localID {
				// update client data
				*c.UIupdate = true; // update UI
				c.getClientList()		 //refresh client list
				c.updateServerRemoteID() //update server
				time.Sleep(100*time.Millisecond)
			}

			// Here we waited and now remote client ID == our remote client ID
			c.l.Info(c.RemoteID+" accepted our connection, performing punch now")

			// Initiate hole punch
			if err := c.pingNpunch(); err != nil {
				c.l.Error(err.Error())
				return
			}
			c.l.Info("Hole punching to "+client)

		}(c)

		return
	}
}

func (c *HPClient) removeFromServer() {
	// Remove our client from server
	// Resolve server adress
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Error resolving server address:"+err.Error())
		return
	}
	// Connect to server
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Error connecting to server:"+err.Error())
		return
	}
	defer conn.Close()
	// Send disconnect request
	request := `{"local_id":"`+c.localID+`","remote_id":"\u0000"}`	// disconnect request
	requestJSON := []byte(request)
	_, err = conn.Write(requestJSON)
	if err != nil {
		c.l.Warn("Error disconnect request:"+err.Error())
		return
	}
}

func (c *HPClient) pingNpunch() error {
	// Run asynchronously in a goroutine
	go func() {
		time.Sleep(time.Second*1)	// time for other client to check server data

		// remove our client from server for clean slate
		c.removeFromServer()

		time.Sleep(2*time.Second)	// time for other client to remove itself from server

		// start pinging code
		c.echo()
	}()
	
	return nil
}

// pinging code
func (c *HPClient) echo() {
	// Reconnect to echo client, by getting address from info exchange server
	r, err := c.REconnect()
	if err != nil {
		c.l.Error("can't connect"+err.Error())
	}

	// Here connection is made, so we can start ping

	var time1, time2 time.Time
	buf := make([]byte, 1024)
	pingCount := 0
	for {
		// send echo to a remote client
		<-time.After(time.Second * 1)
		if _, err := c.Conn.WriteToUDP([]byte(fmt.Sprintf("ping from %v", c.RemoteID)), r); err != nil {
			c.l.Error(err.Error())
		} else {
			time1 = time.Now()
			c.l.Info(fmt.Sprintf("Ping sent to %v@%v:%v", c.RemoteID, c.R.RemoteIP, c.R.RemotePort))
		}

		// read from connection until timeout
		// then send reconnect signal to restore p2p connection
		c.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		n, err := c.Conn.Read(buf)
		if err != nil {
			// Here we failed to recieve back the ping, therefore we call the reconnect function
			c.l.Error(err.Error())
			// on a successful REconnect a new session will be established and
			// the remote client port and address will be updated
			r, err = c.REconnect()
			if err != nil {
				c.l.Error(err.Error())
				break
			}
			time1 = time.Now()
			continue
		}
		// Here the ping succeded and we can print the rountrip time
		time2 = time.Now()
		c.l.Info(fmt.Sprintf("received %v", string(buf[:n])))
		c.l.Info(fmt.Sprintf("Ping: %s", time2.Sub(time1)))
		// Increment ping counter
		pingCount++
		// After 5 pings, we determine the connection as reliable and we can start opening the udp vpn tunnel
		if pingCount == 5 {
			c.l.Info("Completed 5 pings")
			c.l.Info("Ready to open tunnel")
			// Close socket
			c.Conn.Close()
			// Start tunnel
			c.requestOpenTunnel(fmt.Sprintf("%s", time2.Sub(time1)))

			return
		}
	}
}

// Function to tell tunnelman module to open a tunnel
func (c *HPClient) requestOpenTunnel(ping string) {
	// Create tunnel data for the client we want to connect to
	c.TunnelMan.TunClients[c.RemoteID] = tunnel.ClientTunnelData{
		TunnelOn:		false,
		TunnelAddr:		"10.0.0.1",
		TunnelPorts:	make([]int,0),
		Ping:			ping,
		// for udp tunnels
		EndPIP:			c.R.RemoteIP,
		EndPPort:		c.R.RemotePort,
		EndPAPorts:		make([]int,0),
	}
	//? use c.TunnelMan.AddClient(...)
	// Get the  current UDP port which our client is using to send messages to the other client
	localAddr := c.Conn.LocalAddr().(*net.UDPAddr)
	// Call TunnelMan method
	c.TunnelMan.OpenTunnel(c.localID, strconv.Itoa(localAddr.Port), c.RemoteID)
}

// createP2PConnection sends local and remote ID to the server and waits for the response
func (c *HPClient) createP2PConnection() {
	// Error handling
	var err error
	defer func() { c.connectStatus <- err }()

	// Create variables to recieve response
	response := ConnectResponse{}
	request := ConnectRequest{LocalID: c.localID, RemoteID: c.RemoteID}	// normal connect request
	buf := make([]byte, 1024)	// byte buffer to hold raw response

	// Prepare request JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return
	}

	// Resolve exchange server address
	serverUDPAddr, err := net.ResolveUDPAddr("udp",
		c.serverAddr+":"+c.serverPort)
	if err != nil {
		return
	}

	// Listen for echange server response
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return
	}
L:
	for {
		select {
		case <-c.stopChan:
			// if echo channel is stopped, return and exit from endless timeout loop
			return
		default:
			// Set deadline for reading the response from the exchange server
			conn.SetReadDeadline(time.Now().Add(c.timeout + c.timeout/2))
			// Send request to exchange server
			_, err := conn.WriteToUDP(requestJSON, serverUDPAddr)
			if err != nil {
				continue
			}
			// Read raw request buffer
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			// Decode JSON response
			err = json.Unmarshal(buf[:n], &response)
			if err != nil {
				continue
			}

			conn.SetReadDeadline(time.Now().Add(c.timeout + c.timeout/2))
			break L
		}
	}

	// Here we succesfully got the other clients IP and port number from the info exchange server, return function
	c.Conn = conn
	c.R = &response
	return
}

// REconnect allows to manually send reconnect signal, returns new remote addr and port
func (c *HPClient) REconnect() (*net.UDPAddr, error) {
	// call createP2P connection asynchronously using goroutine
	go c.createP2PConnection()
	// listen to channel status for updates
	select {
	case err := <-c.connectStatus:
		if err != nil {
			return nil, err
		}

		// Here the status through the channel has changed to succesfully connected

		// Resolve the adress of the other client
		remoteUDP, err := net.ResolveUDPAddr("udp", c.R.RemoteIP+":"+c.R.RemotePort)
		
		c.l.Info("Got hole-punch addr from exchange server: "+c.RemoteID+"@"+c.R.RemoteIP+":"+c.R.RemotePort)
		
		// Remove self from server client listing
		c.removeFromServer()
		
		// Update the ClientList with the remote IP we just got from the info exchange server
		// First we get a "copy" of the entry
		if entry, ok := c.ClientList[c.RemoteID]; ok {
			// Then we modify the copy
			entry.LocalIP = c.R.RemoteIP
			// Clear all other clients from the map, as at this point of the program we only care about connecting to the client we selected
			c.ClientList = make(map[string]ClientData)
			// Then we reassign map entry
			c.ClientList[c.RemoteID] = entry
		}
		// Update ClientList with the remote port
		// First we get a "copy" of the entry
		if entry, ok := c.ClientList[c.RemoteID]; ok {
			// Then we modify the copy
			entry.LocalPort = c.R.RemotePort
			// Then we reassign map entry
			c.ClientList[c.RemoteID] = entry
		}
		// update UI
		*c.UIupdate = true
		
		return remoteUDP, err

	case <-time.After(time.Second * 60):
		// if it takes over 60 seconds to reconnect, timeout has occured
		c.stopChan <- struct{}{}
		return nil, errors.New("reconnection timeout")
	}
}