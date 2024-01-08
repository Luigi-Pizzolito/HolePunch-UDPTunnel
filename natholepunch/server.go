package natholepunch

import (
	"sync"
	"strconv"
	"encoding/json"
	"fmt"

	"time"
	"net"

	// tui "github.com/Luigi-Pizzolito/HolePunch-UDPTunnel/tui"
	"go.uber.org/zap"
)

// Define the Hole Punch Server struct
type HPServer struct {
	sync.RWMutex
	ClientList map[string]ClientData

	l *zap.Logger

	ConnLogC chan string
}

// Initialises a pointer to a new HPServer struct
func NewHPServer(l *zap.Logger, ConnLogC chan string) *HPServer {
	return &HPServer{ClientList: make(map[string]ClientData), l: l, ConnLogC: ConnLogC}
}

//-------- Server Functions --------
// add client to clients queue map
func (s *HPServer) addClient(c ClientData) {
	s.Lock()
	defer s.Unlock()
	s.ClientList[c.LocalID] = c
}

// start serving and reply to client requests
func (s *HPServer) Serve(serverPort string) error {
	s.l.Info("Info Exchange Server mode")

	// Get server local address bind
	serverAddr, err := net.ResolveUDPAddr("udp", ":"+serverPort)
	if err != nil {
		s.l.Warn("Error", zap.Error(err))
		return err
	}

	// Listen to incoming UDP connections
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		s.l.Warn("Error", zap.Error(err))
		return err
	}
	defer serverConn.Close()

	// s.l.Info("Running server at "+serverAddr.IP.String()+":"+serverPort)
	s.l.Info("Running server at udp://0.0.0.0:"+serverPort)

	// Attend to client connections
	for {
		// Create a buffer to accept new incoming connection
		buf := make([]byte, 1024)
		n, newClientAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			s.l.Warn("Error", zap.Error(err))
			continue
		}

		// Handle each new client in a new goroutine
		go func() {
			// store address of incoming connection
			incomingRequest := ConnectRequest{}
			newClientIP := newClientAddr.IP.String()
			newClientPort := strconv.Itoa(newClientAddr.Port)

			// parse JSON request from client
			json.Unmarshal(buf[:n], &incomingRequest)

			// Add client to clients queue
			s.l.Info("New client: "+incomingRequest.LocalID+"@"+newClientIP+":"+newClientPort)
			s.addClient(ClientData{
				RemoteID:  incomingRequest.RemoteID,
				LocalID:   incomingRequest.LocalID,
				LocalIP:   newClientIP,
				LocalPort: newClientPort,
			})


			// Check if client wants server client list or requests a hole-punch
			if incomingRequest.RemoteID != "" {
				// Client is requesting hole-punch addr of other client

				// Check if other client is online



				s.printPunch(incomingRequest.LocalID,incomingRequest.RemoteID)
				s.l.Info("New client requests hole-punch: "+incomingRequest.LocalID+" -> "+incomingRequest.RemoteID)


			} else {
				// Client is idle and wants list of online clients from server

				// Return list of available online clients

				s.l.Info(incomingRequest.LocalID+" fetched list of clients.")

			}
	
			
		}()
	}

	return nil
}

func (s *HPServer) printPunch(from string, to string) {
	// Get timestampo
	// Format the time to hh:mm:ss
	timeFormatted := time.Now().Format("15:04:05")

	// Print to connection history
	s.ConnLogC <- fmt.Sprintf(" [gray]%s[-] %s [blue]->[-] %s\n", timeFormatted, from, to)
}
