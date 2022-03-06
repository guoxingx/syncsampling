package tcpconn

import (
	"bufio"
	"encoding/json"
	"net"
)

// Session holds connections of server
type Session struct {
	conn      *net.TCPConn
	enc       *json.Encoder
	receiveCh chan []byte
	quitCh    chan struct{}

	ip    string
	mSize int
}

// NewSession returns an session instance
func NewSession(conn *net.TCPConn, mSize int) (*Session, error) {
	s := Session{
		conn:      conn,
		enc:       json.NewEncoder(conn),
		receiveCh: make(chan []byte, 1000),
		quitCh:    make(chan struct{}),
		ip:        conn.RemoteAddr().String(),
		mSize:     mSize,
	}
	return &s, nil
}

// send data to connecion
func (e *Session) send(v interface{}) error {
	if data, ok := v.([]byte); ok {
		data = append(data, []byte("\n")...)
		e.conn.Write(data)
		return nil
	}
	err := e.enc.Encode(v)
	if err != nil {
		// if error occured, close conn
		e.doClose()
	}
	return err
}

// receive data from connection
func (e *Session) receive() <-chan []byte {
	return e.receiveCh
}

// handle connection messages
// func (e *Session) handle(ch chan<- Message) error {
func (e *Session) handle() error {
	// func (e *Session) Handle(ch chan<- []byte) error {
	defer e.doClose()

	connbuf := bufio.NewReaderSize(e.conn, e.mSize)
	for {
		select {
		case <-e.quitCh:
			return nil
		default:
			// message come in from conn
			data, isPrefix, err := connbuf.ReadLine()
			if isPrefix {
				return errPrefixed
			}
			if err != nil {
				return err
			}

			// data as []byte will be reused from the same connbuf
			// due to the principle of slice and array
			// without copy(), consumer data will be mixed with each other
			b := make([]byte, len(data))
			copy(b, data)

			// send message into receive channel
			// ch <- Message{e.ip, b}
			e.receiveCh <- b
		}
	}
}

// close session and connection
func (e *Session) close() error {
	if len(e.quitCh) < cap(e.quitCh) {
		e.quitCh <- struct{}{}
	}
	return nil
}

func (e *Session) doClose() {
	close(e.receiveCh)
	e.conn.Close()
}
