package tcpconn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
)

// Client to request tcp connection
type Client struct {
	config ClientConfig
	host   *net.TCPAddr

	isConnected int32
	sendCh      chan interface{}
	receiveCh   chan []byte
	quitCh      chan struct{}
	infoCh      chan string
	errCh       chan error
	infoOut     chan string
	errOut      chan error
}

// NewClient returns a tcp client instance
func NewClient(cfg ClientConfig) (*Client, error) {
	host, err := net.ResolveTCPAddr("tcp", cfg.Host)
	if err != nil {
		return nil, err
	}

	c := Client{
		config:    cfg,
		host:      host,
		sendCh:    make(chan interface{}, 10000),
		receiveCh: make(chan []byte, 10000),
		quitCh:    make(chan struct{}),
		infoCh:    make(chan string, 1000),
		errCh:     make(chan error, 1000),
		infoOut:   make(chan string, 1000),
		errOut:    make(chan error, 1000),
	}
	return &c, nil
}

// Connect start connect to the host
func (c *Client) Connect() error {
	// return err is already connected
	if atomic.LoadInt32(&c.isConnected) == 1 {
		return errAlreadConnected
	}

	// start conn
	conn, err := net.DialTCP("tcp", nil, c.host)
	if err != nil {
		return err
	}
	defer conn.Close()

	// set status as 1 and handle conn
	atomic.StoreInt32(&c.isConnected, 1)
	defer close(c.receiveCh)
	defer atomic.StoreInt32(&c.isConnected, 0)
	defer conn.Close()

	return c.handleConn(conn)
}

// Disconnect means disconnect
func (c *Client) Disconnect() error {
	// send quit notify
	if atomic.LoadInt32(&c.isConnected) == 0 {
		return fmt.Errorf("not connected")
	}
	c.quitCh <- struct{}{}
	return nil
}

// Send data
func (c *Client) Send(v interface{}) error {
	// return error if not connected
	if atomic.LoadInt32(&c.isConnected) == 0 {
		return errNotConnected
	}

	// send data into sendCh
	c.sendCh <- v
	return nil
}

// IsConnected return false if not connected
func (c *Client) IsConnected() bool {
	return atomic.LoadInt32(&c.isConnected) == 1
}

// Receive returns an output chan of received data
func (c *Client) Receive() <-chan []byte {
	return c.receiveCh
}

// Info returns an output chan of logs
func (c *Client) Info() <-chan string {
	return c.infoOut
}

// Error returns an output chan of logs
func (c *Client) Error() <-chan error {
	return c.errOut
}

func (c *Client) handleConn(conn *net.TCPConn) error {
	// conn must be closed when exit
	// then set status as 0
	go c.handleLogs()

	// encoder for write data into conn
	encoder := json.NewEncoder(conn)
	// connbuf for read data from conn
	connbuf := bufio.NewReaderSize(conn, c.config.MessageSize)
	c.infoCh <- "client handle connection"
	go func() {
		for {
			// message come in from conn
			data, isPrefix, err := connbuf.ReadLine()
			if isPrefix {
				c.errCh <- fmt.Errorf("receive prefix, data: %v", data)
				break
				// return errPrefixed
			}
			if err != nil {
				c.errCh <- fmt.Errorf("connection read error: %s", err)
				break
				// return err
			}

			// send message into receive channel
			b := make([]byte, len(data))
			copy(b, data)
			c.receiveCh <- b
		}
		if atomic.LoadInt32(&c.isConnected) == 1 {
			c.quitCh <- struct{}{}
		}
		c.quitCh <- struct{}{}
	}()

	for {
		select {
		case <-c.quitCh:
			// receive quit notify
			c.infoCh <- "receive quite signal, client will be closed"
			return nil

		case i := <-c.sendCh:
			// receive from send channel
			// send data into conn
			// encoder.Encode(i)
			if data, ok := i.([]byte); ok {
				data = append(data, []byte("\n")...)
				_, err := conn.Write(data)
				if err != nil {
					return err
				}
			} else {
				err := encoder.Encode(i)
				if err != nil {
					return err
				}
			}
		}
	}
}

func (c *Client) handleLogs() {
	for {
		select {
		case i := <-c.infoCh:
			if len(c.infoOut) == cap(c.infoOut) {
				for i := 0; i < cap(c.infoOut)/2; i++ {
					<-c.infoOut
				}
			}
			c.infoOut <- i
		case e := <-c.errCh:
			if len(c.errOut) == cap(c.errOut) {
				for i := 0; i < cap(c.errOut)/2; i++ {
					<-c.errOut
				}
			}
			c.errOut <- e
		}
	}
}
