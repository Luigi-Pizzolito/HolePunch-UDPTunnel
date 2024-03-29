package natholepunch

import (
	"sync"
	"strconv"
	"encoding/json"
	"fmt"
	"errors"

	"time"
	"net"

	"go.uber.org/zap"
)

// Define the Hole Punch Server struct
type HPServer struct {
	// Client list
	sync.RWMutex
	ClientList map[string]ClientData

	// UI update hook
	UIupdate *bool

	// Logger
	l *zap.Logger
	ConnLogC chan string
}

// Initialises a pointer to a new HPServer struct
func NewHPServer(l *zap.Logger, ConnLogC chan string, UIupdate *bool) *HPServer {
	return &HPServer{ClientList: make(map[string]ClientData), l: l, ConnLogC: ConnLogC, UIupdate: UIupdate}
}

//-------- Server Functions --------
// add client to clients queue map
func (s *HPServer) addClient(c ClientData) {
	s.Lock()
	defer s.Unlock()
	s.ClientList[c.LocalID] = c
}
// delete client from clients queue map
func (s *HPServer) deleteClient(ID string) {
	s.RLock()
	defer s.RUnlock()
	delete(s.ClientList, ID)
}
// checkClient checks if client already in the map
func (s *HPServer) checkClient(ID string) (ClientData, error) {
	s.RLock()
	defer s.RUnlock()
	i, ok := s.ClientList[ID]
	if ok {
		return i, nil
	}
	return i, errors.New("no such client")
}

// start serving and reply to client requests
func (s *HPServer) Serve(serverPort string) error {
	s.l.Info("Info Exchange Server mode")

	// Get server local address bind
	serverAddr, err := net.ResolveUDPAddr("udp", ":"+serverPort)
	if err != nil {
		s.l.Warn("Error resolving server UDP addr: "+err.Error())
		return err
	}

	// Listen to incoming UDP connections
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		s.l.Warn("Error listening to UDP connections: "+err.Error())
		return err
	}
	defer serverConn.Close()

	s.l.Info("Running server at udp://0.0.0.0:"+serverPort)

	// Attend to client connections
	for {
		// Create a buffer to accept new incoming connection
		buf := make([]byte, 1024)
		n, newClientAddr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			s.l.Warn("Error reading UDP packet from client request: "+err.Error())
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

			// Here there are 3 types of requests which can be accepted
			// the request type is defined by the content of incomingRequest.RemoteID
			// ---- RemoteID string: 	connection to other client hole-punch request
			// ---- "" string: 			get online client listing from server (idle client)
			// ---- "\x00" string:		disconnect idle client from server

			// Client disconnect request
			if incomingRequest.RemoteID == "\x00" {
				// Client was idle and is now closing the program
				// Remove client from list of available online clients
				s.deleteClient(incomingRequest.LocalID)
				s.l.Info(incomingRequest.LocalID+" disconnected.")

				*s.UIupdate = true; // update UI

				return
			} else {
				// Other kind of request: Idle listing or Hole-Punch request

				// Add client to clients queue
				// check if client is not already in queue first
				s.RLock()
				if _, ok := s.ClientList[incomingRequest.LocalID]; !ok {
					s.l.Info("New client: "+incomingRequest.LocalID+"@"+newClientIP+":"+newClientPort)
					*s.UIupdate = true; // update UI
				}
				s.RUnlock()
				// Add client to queue
				s.addClient(ClientData{
					RemoteID:  incomingRequest.RemoteID,
					LocalID:   incomingRequest.LocalID,
					LocalIP:   newClientIP,
					LocalPort: newClientPort,
				})
				

				// Check if client wants server client list or requests a hole-punch
				if incomingRequest.RemoteID == "" {
					// Client is idle and wants list of online clients from server

					// Build response to client JSON
					// create copy of client list and remove requester LocalID, can't connect to yourself!
					s.RLock()
					clientList := make(map[string]ClientData, len(s.ClientList))
					for k, v := range s.ClientList {
						clientList[k] = v
					}
					s.RUnlock()
					delete(clientList, incomingRequest.LocalID)

					// Convert to JSON
					responseToClientListingJSON, err := json.Marshal(clientList)
					if err != nil {
						s.l.Warn("Error preparing client listing response JSON: "+err.Error())
						return
					}

					// Send response listing to client
					_, err = serverConn.WriteToUDP(responseToClientListingJSON, newClientAddr)
					if err != nil {
						s.l.Warn("Error sending client listing response JSON: "+err.Error())
						return
					}

					// s.l.Info(incomingRequest.LocalID+" fetched list of clients.")

					return
				} else {
					// Client is requesting hole-punch addr of other client

					clientFromMap, err := s.checkClient(incomingRequest.RemoteID)
					// if requested RemoteID is not available (nil), client is in the waiting state until the requested RemoteID becomes online

					//! or if the LocalID is not the requested RemoteID (remote ID also needs to accept/initiate the connection, 2way)
				
					if err != nil || clientFromMap.RemoteID != incomingRequest.LocalID {
						// s.l.Info(incomingRequest.LocalID+" is waiting for "+incomingRequest.RemoteID)
						// update waiting status
						s.RLock()
						// s.ClientList[incomingRequest.LocalID].RemoteID = incomingRequest.RemoteID;
						// First we get a "copy" of the entry
						if entry, ok := s.ClientList[incomingRequest.LocalID]; ok {
							// Then we modify the copy
							entry.RemoteID = incomingRequest.RemoteID
							s.RUnlock()
							s.Lock()
							// Then we reassign map entry
							s.ClientList[incomingRequest.LocalID] = entry
							s.Unlock()
						}

						*s.UIupdate = true; // update UI
						
						// exit handle function
						return
					}

					// If we have made it here, then the requested RemoteID is online and we can reply with its UDP address for hole punching

					// Build response to client JSON (goes to original client that requested hole-punch)
					responseToClientFromMap := ConnectResponse{
						RemoteIP:	clientFromMap.LocalIP,
						RemotePort:	clientFromMap.LocalPort,
					}
					responseToClientFromMapJSON, err := json.Marshal(responseToClientFromMap)
					if err != nil {
						s.l.Warn("Error preparing client response JSON: "+err.Error())
						return
					}

					// Build response to new client JSON (goes to destination client to perform NAT hole-punching)
					responseToPunchedClient := ConnectResponse{
						RemoteIP:	newClientIP,
						RemotePort:	newClientPort,
					}
					responseToPunchedClientJSON, err := json.Marshal(responseToPunchedClient)
					if err != nil {
						s.l.Warn("Error preparing punched client response JSON: "+err.Error())
						return
					}

					// Send response to original requester client
					_, err = serverConn.WriteToUDP(responseToClientFromMapJSON, newClientAddr)
					if err != nil {
						s.l.Warn("Error sending client response JSON: "+err.Error())
						return
					}

					// Get UDP addr of punched client
					clientFromMapUDPAddr, _ := net.ResolveUDPAddr("udp", clientFromMap.LocalIP+":"+clientFromMap.LocalPort)
					if err != nil {
						s.l.Warn("Error resolving punched client UDP addr: "+err.Error())
						return
					}

					// Send UDP response to punched client NAT to perform hole-punch
					_, err = serverConn.WriteToUDP(responseToPunchedClientJSON, clientFromMapUDPAddr)
					if err != nil {
						s.l.Warn("Error sending punch client punching response: "+err.Error())
						return
					}

					// At this point client has recieved all the information it needs to perform hole-punch!

					// Print succesfull punch request to logs
					s.printPunch(incomingRequest.LocalID,incomingRequest.RemoteID)
					s.l.Info("Punch request: "+incomingRequest.LocalID+" -> "+incomingRequest.RemoteID)
				
				}
			}
		}()
	}

	return nil
}

// Print punch into connection history on the right side of the server TUI
func (s *HPServer) printPunch(from string, to string) {
	// Get timestampo
	// Format the time to hh:mm:ss
	timeFormatted := time.Now().Format("15:04:05")

	// Print to connection history
	s.ConnLogC <- fmt.Sprintf(" [gray]%s[-] %s [blue]->[-] %s\n", timeFormatted, from, to)
}
