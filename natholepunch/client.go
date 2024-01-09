package natholepunch

import (
	"encoding/json"
	"time"
	"net"
	"fmt"
	"errors"
	"bytes"
	// "strconv"

	// tunnel "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/udptunnel"

	"go.uber.org/zap"

	// "bufio"
	// "io"
	// "os"
	// "os/exec"
	// "path/filepath"
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
	//todo: Tunnel info & udptunnel command channel
	//TunnelInfo		map...
	//TunnelControl		chan ?
	// Logger
	l				*zap.Logger
}

// Initialises a pointer to a new HPClient struct
func NewHPClient(l *zap.Logger, timeout time.Duration, serverAddr, serverPort, localID, remoteID string, UIupdate, ClientUIStage2 *bool,/*, bind UI tunnel listing here later*/) *HPClient {
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
	}
}

// Main logic
func (c *HPClient) Run() error {
	c.l.Info("Hole-Punch & UDP Tunnel Client mode")

	// Sync online clients list from server
	go func(){
		for {
			if !c.pauseClientFetch {
				c.getClientList()
				//? maybe need to add a channel here to pause this when creating tunnel
			}
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
	
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Client List: Error resolving server address:"+err.Error())
		return err
	}

	// laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	// if err != nil {
	// 	c.l.Warn("Client List: Error resolving local UDP send address:"+err.Error())
	// 	return err
	// }

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Client List: Error connecting to server:"+err.Error())
		return err
	}
	defer conn.Close()

	request := `{"local_id":"`+c.localID+`","remote_id":""}`	// client listing request when idle as c.RemoteID=="", otherwise update with waiting
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

	// c.l.Warn(strconv.FormatBool(bytes.Equal(c.UIupdateCListB, buffer)))
	if !bytes.Equal(c.UIupdateCListB, buffer) {
		// if the response is different, trigger an UI update
		*c.UIupdate = true;
		// c.l.Warn("UI update from getClientList")
		// c.l.Warn(strconv.FormatBool(bytes.Equal(c.UIupdateCListB, buffer)))
		copy(c.UIupdateCListB, buffer)
		// c.l.Warn(strconv.FormatBool(bytes.Equal(c.UIupdateCListB, buffer)))
	}

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
	// c.l.Info("Updating server client status to waiting for "+c.RemoteID)
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
	if err != nil {
		c.l.Warn("Client List: Error resolving server address:"+err.Error())
		return err
	}

	// laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	// if err != nil {
	// 	c.l.Warn("Client List: Error resolving local UDP send address:"+err.Error())
	// 	return err
	// }

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		c.l.Warn("Client List: Error connecting to server:"+err.Error())
		return err
	}
	defer conn.Close()

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

	*c.ClientUIStage2 = true

	c.l.Info("Request for hole punch to "+client)

	// set our clients remoteID
	c.pauseClientFetch = true;
	c.RemoteID = client;
	// update server with remoteID for connection request
	c.updateServerRemoteID()
	*c.UIupdate = true; // update UI
	

	// check if we are ready to hole punch (other client also wants)
	clientData := c.ClientList[client]
	if clientData.RemoteID == c.localID {
		c.l.Info("Remote client also wants to connect")
		c.l.Info("Initiating hole punch to "+client)
		if err := c.pingNpunch(); err != nil {
			c.l.Error(err.Error())
			return
		}
		c.l.Info("Hole punching to "+client)

		

		// return to idle
		// c.RemoteID = "";	//update local
		// *c.UIupdate = true; // update fUI
		// c.pauseClientFetch = false;
		// c.updateServerRemoteID()	//update server

	} else {
		c.l.Warn("Remote client does not want to connect") 
		if clientData.RemoteID == "" {
			c.l.Info("Remote client is Idle")
		} else {
			c.l.Info("Remote client is waiting for "+clientData.RemoteID)
		}
		c.l.Info("Please ask remote client to connect to you too")

		c.pauseClientFetch = true;
		go func(c *HPClient) {
			c.l.Info("Waiting in loop for "+c.RemoteID)
			for c.ClientList[c.RemoteID].RemoteID != c.localID {
				// update client data
				*c.UIupdate = true; // update UI
				// c.pauseClientFetch = false;
				c.getClientList()		 //refresh client list
				c.updateServerRemoteID() //update server
				time.Sleep(100*time.Millisecond)
				// c.pauseClientFetch = true;
			}
			// Here we waited and now remote client ID == our remote client ID
			c.l.Info(c.RemoteID+" accepted our connection, performing punch now")

			// c.InitiatePunch(c.RemoteID)
			if err := c.pingNpunch(); err != nil {
				c.l.Error(err.Error())
				return
			}
			c.l.Info("Hole punching to "+client)


		}(c)
		
		// return to idle
		// c.RemoteID = "";	//update local
		// *c.UIupdate = true; // update UI
		// c.pauseClientFetch = false;
		// c.updateServerRemoteID()	//update server

		return
	}
}

func (c *HPClient) removeFromServer() {
	// remove our client from server
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr+":"+c.serverPort)
		if err != nil {
			c.l.Warn("Error resolving server address:"+err.Error())
			return
		}
		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			c.l.Warn("Error connecting to server:"+err.Error())
			return
		}
		defer conn.Close()
		request := `{"local_id":"`+c.localID+`","remote_id":"\u0000"}`	// disconnect request
		requestJSON := []byte(request)
		_, err = conn.Write(requestJSON)
		if err != nil {
			c.l.Warn("Error disconnect request:"+err.Error())
			return
		}
}

func (c *HPClient) pingNpunch() error {
	// go c.echo()
	go func() {

		

		time.Sleep(time.Second*1)	// time for other client to check server data

		// remove our client from server for clean slate
		c.removeFromServer()

		time.Sleep(2*time.Second)	// time for other client to remove itself from server

		

		// start original punch n' echo code
		// runCommand(c.l, "./natholepunch/udp-nat-hole-punch-exe", "-l", c.localID, "-r", c.RemoteID)
		c.echo()
	}()
	/*
	// Punch
	r, err := c.REconnect()
	if err != nil {
		return errors.New("Can't connect to client "+c.RemoteID)
	}

	// Ping
	var time1, time2 time.Time
	pingTimeout := 0
	buf := make([]byte, 1024)
	for {
		<- time.After(2*time.Second)
		// if pingTimeout > 5 {
		// 	return errors.New("ping to udp://"+c.RemoteID+"@"+c.R.RemoteIP+":"+c.R.RemotePort+" timed out 5 times, giving up.")
		// }

		// send echo to a remote client
		if _, err := c.Conn.WriteToUDP([]byte(fmt.Sprintf("ping from %v", c.localID)), r); err != nil {
			return errors.New("Error sending ping:"+err.Error())
		} else {
			time1 = time.Now()
			c.l.Info("Ping sent to udp://"+c.RemoteID+"@"+c.R.RemoteIP+":"+c.R.RemotePort)
		}
		// read from connection until timeout
		// then send reconnect signal to restore p2p connection
		c.Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		n, err := c.Conn.Read(buf)
		if err != nil {
			c.l.Error("Timeout recieving ping, retrying:"+err.Error())
			// on a successful REconnect a new session will be established and
			// the remote client port and address will be updated
			r, err = c.REconnect()
			if err != nil {
				c.l.Error("Failed to reconnect to "+c.RemoteID+" after ping timeout, retrying:"+err.Error())
			}
			pingTimeout++
			continue
		}
		time2 = time.Now()
		c.l.Info("Received "+string(buf[:n]))
		c.l.Info("Ping: "+time2.Sub(time1).String())
		// return nil
	}
	*/
	return nil
}

// func runCommand(l *zap.Logger, command string, args ...string) {
// 	// Get the path of the currently running executable
// 	executablePath, err := os.Executable()
// 	if err != nil {
// 		fmt.Println("Error getting executable path:", err)
// 		return
// 	}

// 	// Get the directory of the executable
// 	executableDir := filepath.Dir(executablePath)

// 	// Change the working directory to the directory of the executable
// 	err = os.Chdir(executableDir)
// 	if err != nil {
// 		fmt.Println("Error changing working directory:", err)
// 		return
// 	}

// 	cmd := exec.Command(command, args...)

// 	stdoutPipe, _ := cmd.StdoutPipe()
// 	stderrPipe, _ := cmd.StderrPipe()

// 	err = cmd.Start()
// 	if err != nil {
// 		fmt.Println("Error starting command:", err)
// 		return
// 	}

// 	go relayOutput(l, stdoutPipe, "STDOUT")
// 	go relayOutput(l, stderrPipe, "STDERR")

// 	err = cmd.Wait()
// 	if err != nil {
// 		fmt.Println("Error waiting for command:", err)
// 	}
// }

// func relayOutput(l *zap.Logger, pipe io.Reader, label string) {
// 	scanner := bufio.NewScanner(pipe)
// 	for scanner.Scan() {
// 		l.Info(fmt.Sprintf("[%s] %s", label, scanner.Text()))
// 	}
// }

func (c *HPClient) echo() {

	r, err := c.REconnect()
	if err != nil {
		c.l.Error("can't connect"+err.Error())
	}

	

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
		time2 = time.Now()
		c.l.Info(fmt.Sprintf("received %v", string(buf[:n])))
		c.l.Info(fmt.Sprintf("Ping: %s", time2.Sub(time1)))

		pingCount++
		
		if pingCount == 5 {
			c.l.Info("Completed 5 pings")
			c.l.Info("Ready to open tunnel")
			//!added: here the hole punch worked and we have already sent a bidirectional ping
			return
		}
	}
}

// createP2PConnection sends local and remote ID to the server and waits for the response
func (c *HPClient) createP2PConnection() {
	var err error
	defer func() { c.connectStatus <- err }()

	response := ConnectResponse{}
	request := ConnectRequest{LocalID: c.localID, RemoteID: c.RemoteID}	// normal connect request

	buf := make([]byte, 1024)

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return
	}

	serverUDPAddr, err := net.ResolveUDPAddr("udp",
		c.serverAddr+":"+c.serverPort)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return
	}
L:
	for {
		select {
		case <-c.stopChan:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(c.timeout + c.timeout/2))
			_, err := conn.WriteToUDP(requestJSON, serverUDPAddr)
			if err != nil {
				continue
			}

			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}
			err = json.Unmarshal(buf[:n], &response)
			if err != nil {
				continue
			}

			conn.SetReadDeadline(time.Now().Add(c.timeout + c.timeout/2))
			break L
		}
	}

	c.Conn = conn
	c.R = &response
	return
}

// REconnect allows to manually send reconnect signal, returns new remote addr and port
func (c *HPClient) REconnect() (*net.UDPAddr, error) {
	go c.createP2PConnection()

	select {
	case err := <-c.connectStatus:
		if err != nil {
			return nil, err
		}
		remoteUDP, err := net.ResolveUDPAddr("udp", c.R.RemoteIP+":"+c.R.RemotePort)
		// fmt.Printf("Got P2P IP from exchange server: %v:%v\n", c.R.RemoteIP, c.R.RemotePort)
		
		c.l.Info("Got hole-punch addr from exchange server: "+c.RemoteID+"@"+c.R.RemoteIP+":"+c.R.RemotePort)
		//! added: remove self from server after succesful reconnection
		c.removeFromServer()
		//! added: update ip and ports for TUI
		// update IP
		// First we get a "copy" of the entry
		if entry, ok := c.ClientList[c.RemoteID]; ok {
			// Then we modify the copy
			entry.LocalIP = c.R.RemoteIP
			//!added: clear map here to remove other clients
			c.ClientList = make(map[string]ClientData)
			// Then we reassign map entry
			c.ClientList[c.RemoteID] = entry
		}
		// update port
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
		c.stopChan <- struct{}{}
		return nil, errors.New("reconnection timeout")
	}
}



//! both clients need to request connection for hole punch to work?? check in network test

//!on disconnect client must send connection request with null-byte as remoteID