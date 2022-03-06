package tcpconn

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

// Server handle tcp connnections
type Server struct {
	config ServerConfig
	listen *net.TCPAddr

	isStarted    int32
	sendCh       chan sendReq
	newConnCh    chan string
	receiveRawCh chan []byte
	quitCh       chan struct{}
	infoCh       chan string
	errCh        chan error
	infoOut      chan string
	errOut       chan error

	sessions  map[*Session]struct{}
	ips       map[string]*Session
	sessionMu sync.RWMutex
}

// NewServer returns a tcp server instance
func NewServer(cfg ServerConfig) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", cfg.Listen)
	if err != nil {
		return nil, err
	}

	s := Server{
		config:       cfg,
		listen:       addr,
		sendCh:       make(chan sendReq, 10000),
		newConnCh:    make(chan string, 1000),
		receiveRawCh: make(chan []byte, 10000),
		quitCh:       make(chan struct{}),
		infoCh:       make(chan string, 1000),
		errCh:        make(chan error, 1000),
		infoOut:      make(chan string, 1000),
		errOut:       make(chan error, 1000),
		sessions:     make(map[*Session]struct{}),
		ips:          make(map[string]*Session),
	}

	return &s, nil
}

// Start tcp server
func (s *Server) Start() error {
	// reutrn error if already started
	if atomic.LoadInt32(&s.isStarted) == 1 {
		return errServerAlreadStarted
	}

	return s.listenTCP()
}

// Stop tcp server
func (s *Server) Stop() error {
	// reutrn error if not startd
	if atomic.LoadInt32(&s.isStarted) == 0 {
		return errServerNotStarted
	}

	s.quitCh <- struct{}{}
	return nil
}

// SendAll data to all connected
func (s *Server) SendAll(v interface{}) error {
	// reutrn error if not startd
	if atomic.LoadInt32(&s.isStarted) == 0 {
		return errServerNotStarted
	}

	s.sendCh <- sendReq{v, nil}
	return nil
}

// Send to sessions
func (s *Server) Send(v interface{}, ips []string) error {
	if atomic.LoadInt32(&s.isStarted) == 0 {
		return errServerNotStarted
	}

	s.sendCh <- sendReq{v, ips}
	return nil
}

// Receive returns an output chan of received data
func (s *Server) Receive(ip string) <-chan []byte {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	for e := range s.sessions {
		if e.ip == ip {
			return e.receive()
		}
	}
	return nil
}

// NewConnection ...
func (s *Server) NewConnection() <-chan string {
	return s.newConnCh
}

// Connections returns informations of all connected client
func (s *Server) Connections() []string {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	res := make([]string, 0)
	for e := range s.sessions {
		res = append(res, e.ip)
	}
	return res
}

// Disconnect a connection
func (s *Server) Disconnect(ip string) {
	s.sessionMu.Lock()
	e, ok := s.ips[ip]
	if ok {
		s.sessionMu.Unlock()
		s.removeSession(e, true, true)
	} else {
		s.sessionMu.Unlock()
	}
}

// Info returns an output chan of logs
func (s *Server) Info() <-chan string {
	return s.infoOut
}

// Error returns an output chan of logs
func (s *Server) Error() <-chan error {
	return s.errOut
}

func (s *Server) listenTCP() error {
	server, err := net.ListenTCP("tcp", s.listen)
	if err != nil {
		return err
	}
	defer atomic.StoreInt32(&s.isStarted, 0)
	defer server.Close()

	go s.handleLogs()

	waitCh := make(chan struct{}, s.config.MaxConn)
	atomic.StoreInt32(&s.isStarted, 1)
	s.infoCh <- "server start listening"
	go func() {
		for {
			// handle tcp connections
			conn, err := server.AcceptTCP()
			if err != nil {
				s.errCh <- fmt.Errorf("accept tcp error: %s", err)
				continue
			}
			// blocked if connections reached MaxConn
			waitCh <- struct{}{}
			s.infoCh <- fmt.Sprintf("connection incoming, current: %d", len(waitCh))

			conn.SetKeepAlive(true)
			e, _ := NewSession(conn, s.config.MessageSize)
			s.addSession(e)

			// handle session
			go func(w chan struct{}) {
				err := e.handle()
				if err != nil {
					// err occoured means session was cloesed by accident intead of a server call
					// so do removeSession is necessary
					s.removeSession(e, true, false)
					s.errCh <- fmt.Errorf("connection %s exit with error: %s", e.ip, err)
				}
				conn.Close()

				// release
				<-waitCh
				s.infoCh <- fmt.Sprintf("connection exit, current connected: %d", len(w))
			}(waitCh)
		}
	}()

	for {
		select {
		case <-s.quitCh:
			// receive quite notify
			s.infoCh <- "receive quite signal, server will be closed"
			return nil

		case req := <-s.sendCh:
			// send data to each session
			s.send(req)
		}
	}
}

// send data into sessions
func (s *Server) send(req sendReq) {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()

	var wg sync.WaitGroup
	if req.ips == nil {
		for e := range s.sessions {
			wg.Add(1)
			go s.sessionSend(e, req.v, &wg)
		}
	} else {
		for _, ip := range req.ips {
			e, ok := s.ips[ip]
			if !ok {
				s.errCh <- fmt.Errorf("failed to send to %s", ip)
			} else {
				wg.Add(1)
				go s.sessionSend(e, req.v, &wg)
			}
		}
	}

	wg.Wait()
}

func (s *Server) sessionSend(e *Session, v interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	err := e.send(v)
	if err != nil {
		// if error occured, remove session
		// assume conn is alread closed
		s.errCh <- fmt.Errorf("failed to send, error: %s", err)
		// mu already locked during send
		s.removeSession(e, false, true)
	}
}

func (s *Server) addSession(e *Session) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.sessions[e] = struct{}{}
	s.ips[e.ip] = e
	s.newConnCh <- e.ip
	s.infoCh <- fmt.Sprintf("session add, current: %d", len(s.sessions))
}

func (s *Server) removeSession(e *Session, lock, clos bool) {
	if lock {
		s.sessionMu.Lock()
		defer s.sessionMu.Unlock()
	}
	delete(s.sessions, e)
	delete(s.ips, e.ip)
	if clos {
		e.close()
	}
	s.infoCh <- fmt.Sprintf("session remove, current: %d", len(s.sessions))
}

func (s *Server) handleLogs() {
	for {
		select {
		case i := <-s.infoCh:
			if len(s.infoOut) == cap(s.infoOut) {
				for i := 0; i < cap(s.infoOut)/2; i++ {
					<-s.infoOut
				}
			}
			s.infoOut <- i
		case e := <-s.errCh:
			if len(s.errOut) == cap(s.errOut) {
				for i := 0; i < cap(s.errOut)/2; i++ {
					<-s.errOut
				}
			}
			s.errOut <- e
		}
	}
}
