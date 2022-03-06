package tcpconn

import (
	"errors"
)

var (
	errNotConnected    = errors.New("not connected")
	errPrefixed        = errors.New("prefix is not allowed")
	errAlreadConnected = errors.New("already connected")

	errServerAlreadStarted = errors.New("tcp server already started")
	errServerNotStarted    = errors.New("tcp server not started")
)
