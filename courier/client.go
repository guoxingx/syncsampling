package courier

import (
	"fmt"
	"sync/atomic"

	"syncsampling/utils/tcpconn"
)

// Client is the wrap of tcpconn.Client with Courier
type Client struct {
	*Courier

	tc *tcpconn.Client
}

// NewClientWithTCP ...
func NewClientWithTCP(cfg tcpconn.ClientConfig) (*Client, error) {
	tc, err := tcpconn.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	rawSendCh := make(chan interface{}, 1000)
	rawReceiveCh := tc.Receive()
	c, err := NewCourier(rawSendCh, rawReceiveCh)
	if err != nil {
		return nil, err
	}
	w := Client{c, tc}
	return &w, nil
}

// Run courier
func (w *Client) Run() error {
	if atomic.LoadInt32(&w.running) == 1 {
		return fmt.Errorf("courier client already running")
	}
	defer close(w.receiveCh)
	defer atomic.StoreInt32(&w.running, 0)
	atomic.StoreInt32(&w.running, 1)

	tcpQuit := make(chan error, 2)

	go func(q chan error) {
		err := w.tc.Connect()
		q <- err
	}(tcpQuit)

	for {
		select {
		case v := <-w.rawSendCh:
			w.tc.Send(v)

		case data, ok := <-w.rawReceiveCh:
			if !ok {
				return fmt.Errorf("rawReceiveCh closed")
			}
			m, err := w.decode(data)
			p := Package{
				M:     m,
				Error: err,
			}
			w.receiveCh <- p

		case e := <-tcpQuit:
			// quit signal from tcp
			return e

		case <-w.quitCh:
			// quit signal
			w.tc.Disconnect()
			return nil
		}
	}
}

// Send message
func (w *Client) Send(m interface{}) error {
	data, err := w.encode(m)
	if err != nil {
		return err
	}

	return w.tc.Send(data)
}

// IsConnected ...
func (w *Client) IsConnected() bool {
	return w.tc.IsConnected()
}
