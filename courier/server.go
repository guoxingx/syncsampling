package courier

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"syncsampling/utils/tcpconn"
)

// Server is the wrap of tcpconn.Server with Courier
type Server struct {
	*Courier

	ts     *tcpconn.Server
	newOut chan string
	outs   map[string]chan Package
	outMu  sync.Mutex
}

// NewServerWithTCP ...
func NewServerWithTCP(cfg tcpconn.ServerConfig) (*Server, error) {
	ts, err := tcpconn.NewServer(cfg)
	if err != nil {
		return nil, err
	}

	c := Courier{
		registered: make(map[string]reflect.Type),
		tags:       make(map[string]uint16),
		names:      make(map[uint16]string),
		rawSendCh:  make(chan interface{}),
		quitCh:     make(chan struct{}),
	}

	s := Server{
		Courier: &c,
		ts:      ts,
		newOut:  make(chan string),
		outs:    make(map[string]chan Package),
	}
	return &s, nil
}

// Run courier
func (s *Server) Run() error {
	if atomic.LoadInt32(&s.running) == 1 {
		return fmt.Errorf("courier server already running")
	}

	defer atomic.StoreInt32(&s.running, 0)
	atomic.StoreInt32(&s.running, 1)

	quitCh := make(chan error)

	go func(q chan error) {
		err := s.ts.Start()
		q <- err
	}(quitCh)

	go func() {
		for {
			ip := <-s.ts.NewConnection()
			ch := s.ts.Receive(ip)
			go func(ip string, ch <-chan []byte) {
				defer s.removeOut(ip)
				out := make(chan Package, cap(ch))
				s.addOut(ip, out)
				s.newOut <- ip
				for {
					data, ok := <-ch
					if !ok {
						return
					}
					m, err := s.decode(data)
					p := Package{
						M:     m,
						From:  ip,
						Error: err,
					}
					out <- p
				}
			}(ip, ch)
		}
	}()

	for {
		e := <-quitCh
		return e
	}
}

// Send message
func (s *Server) Send(m interface{}, ips []string) error {
	if atomic.LoadInt32(&s.running) == 0 {
		return fmt.Errorf("courier not running")
	}

	data, err := s.encode(m)
	if err != nil {
		return err
	}

	return s.ts.Send(data, ips)
}

// Receive message with address info
func (s *Server) Receive(ip string) <-chan Package {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	out, _ := s.outs[ip]
	return out
}

// NewConnection ...
func (s *Server) NewConnection() <-chan string {
	return s.newOut
}

// Connections returns ip of all connected
func (s *Server) Connections() []string {
	return s.ts.Connections()
}

// Disconnect a connection
func (s *Server) Disconnect(ip string) {
	s.ts.Disconnect(ip)
}

func (s *Server) addOut(ip string, ch chan Package) {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	s.outs[ip] = ch
}

func (s *Server) removeOut(ip string) {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	ch, ok := s.outs[ip]
	if ok {
		delete(s.outs, ip)
		close(ch)
	}
}
