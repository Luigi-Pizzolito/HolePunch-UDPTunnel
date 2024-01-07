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

// Initialiser
func (s *HPServer) init() {
	s.l.Info("nat server started!")
}

//-------- Server Functions --------
func (s *HPServer) Hello() {
	s.l.Info("Hello from HPServer!")
	// for i := 0; i < 5000000; i++ {
    //     // Display integer.
    //     s.l.Info(strconv.Itoa(i))
	// 	time.Sleep(1 * time.Second)
    // }
}

