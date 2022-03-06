// Package courier set up rules of transmit messages between services
// client and server wraps courier with tcp connection
package courier

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// Courier carry messages and types
type Courier struct {
	// registered map[string]interface{}
	registered map[string]reflect.Type
	tags       map[string]uint16
	names      map[uint16]string
	mu         sync.Mutex

	rawSendCh    chan interface{}
	rawReceiveCh <-chan []byte
	receiveCh    chan Package

	quitCh  chan struct{}
	running int32
}

// NewCourier returns a Courier instance
func NewCourier(rawSendCh chan interface{}, rawReceiveCh <-chan []byte) (*Courier, error) {
	if rawSendCh == nil || rawReceiveCh == nil {
		return nil, fmt.Errorf("send or receive chan can not be nil")
	}
	c := Courier{
		registered:   make(map[string]reflect.Type),
		tags:         make(map[string]uint16),
		names:        make(map[uint16]string),
		rawSendCh:    rawSendCh,
		rawReceiveCh: rawReceiveCh,
		receiveCh:    make(chan Package, cap(rawReceiveCh)),
		quitCh:       make(chan struct{}),
	}
	return &c, nil
}

// Register a a message type
func (c *Courier) Register(m interface{}, tag uint16) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.names[tag]; ok {
		panic(fmt.Errorf("tag %d already registered with %s", tag, n))
	}

	rtp := reflect.TypeOf(m)
	if rtp.Kind() != reflect.Ptr {
		panic(fmt.Errorf("type must be *struct not '%s (*%s)'", rtp, rtp.Kind()))
	}

	name := rtp.Elem().Name()
	if _, ok := c.registered[name]; ok {
		return fmt.Errorf("%s already registered", name)
	}

	// fmt.Printf("register: %d: %s: %T\n", tag, name, m)
	c.registered[name] = rtp
	c.tags[name] = tag
	c.names[tag] = name
	return nil
}

// Send message
func (c *Courier) Send(m interface{}) error {
	if atomic.LoadInt32(&c.running) == 0 {
		return fmt.Errorf("courier not running")
	}

	data, err := c.encode(m)
	if err != nil {
		return err
	}

	c.rawSendCh <- data
	return nil
}

// Receive message with address info
func (c *Courier) Receive() <-chan Package {
	return c.receiveCh
}

// Run courier
func (c *Courier) Run() error {
	if atomic.LoadInt32(&c.running) == 1 {
		return fmt.Errorf("courier already running")
	}

	defer close(c.receiveCh)
	defer atomic.StoreInt32(&c.running, 0)
	atomic.StoreInt32(&c.running, 1)

	for {
		select {
		case b, ok := <-c.rawReceiveCh:
			if !ok {
				return fmt.Errorf("raw receive chan closed")
			}
			m, err := c.decode(b)
			p := Package{
				M:     m,
				Error: err,
			}
			c.receiveCh <- p
		case <-c.quitCh:
			return nil
		}
	}
}

// Close courier
func (c *Courier) Close() error {
	if atomic.LoadInt32(&c.running) == 0 {
		return fmt.Errorf("already closed")
	}
	c.quitCh <- struct{}{}
	return nil
}

func (c *Courier) encode(m interface{}) ([]byte, error) {
	rtp := reflect.TypeOf(m)
	if rtp.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("type must be *struct not '%s (*%s)'", rtp, rtp.Kind())
	}
	name := rtp.Elem().Name()
	tag, ok := c.tags[name]
	if !ok {
		return nil, fmt.Errorf("unregisterd message type: %s", name)
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(make([]byte, 0, 2+len(data)))
	binary.Write(buff, binary.BigEndian, tag)
	buff.Write(data)
	return buff.Bytes(), nil
}

func (c *Courier) decode(b []byte) (interface{}, error) {
	t := b[:2]
	data := b[2:]

	tag := binary.BigEndian.Uint16(t)
	name, ok := c.names[tag]
	if !ok {
		return nil, fmt.Errorf("unregisterd message type, data: %v", b)
	}
	rtp := c.registered[name]
	rvp := reflect.New(rtp.Elem())
	v := rvp.Interface()

	err := json.Unmarshal(data, &v)
	return v, err
}
