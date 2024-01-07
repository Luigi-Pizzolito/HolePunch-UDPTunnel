package natholepunch

import (
	"sync"
	// "strconv"

	// "time"

	"go.uber.org/zap"
)

// Define the Hole Punch Server struct
type HPServer struct {
	sync.RWMutex
	m map[string]ClientData
	l *zap.Logger
}

// Initialises a pointer to a new HPServer struct
func NewHPServer(l *zap.Logger) *HPServer {
	return &HPServer{m: make(map[string]ClientData), l: l}
}

//-------- Server Functions --------
func (s *HPServer) Serve(serverAddr string, serverPort string) {
	s.l.Info("Info Exchange Server mode")
	s.l.Info("Running server at "+serverAddr+":"+serverPort)
}

