package transmitter

import (
	"sync/atomic"

	"go.uber.org/zap"

	"syncsampling/courier"
	"syncsampling/logs"
	"syncsampling/utils/tcpconn"
	"syncsampling/webserver"
)

var (
	tcp    tcpconn.ServerConfig
	logger *zap.SugaredLogger
)

type Transmitter struct {
	tcp *courier.Server

	rxIP    string
	txIndex int64
	rxIndex int64
}

func NewTransmitter() (*Transmitter, error) {
	logger = logs.GetLogger()

	tcp = tcpconn.ServerConfig{
		Listen:      "127.0.0.1:26001",
		MaxConn:     10,
		MessageSize: 65535,
	}

	server, err := courier.NewServerWithTCP(tcp)
	if err != nil {
		return nil, err
	}
	registerMessage(server)

	m := Transmitter{
		tcp:     server,
		txIndex: 0,
		rxIndex: -1,
	}

	// reset webserver.Index to 0
	atomic.StoreInt32(&webserver.Index, 0)

	return &m, nil
}

func (tx *Transmitter) Run() error {
	// start tcp server
	go func() {
		logger.Infof("start tcp server")
		err := tx.tcp.Run()
		if err != nil {
			logger.Fatalf("tcp server exit: %s", err)
		}
	}()

	// start webserver
	go func() {
		logger.Infof("start web server")
		err := webserver.StartServer()
		if err != nil {
			logger.Fatalf("web server exit: %s", err)
		}
	}()

	return tx.handleTCP()
}

func (tx *Transmitter) handleTCP() error {
	for {
		logger.Info("tcp server ready")

		// new tcp connection
		ip := <-tx.tcp.NewConnection()
		ch := tx.tcp.Receive(ip)
		logger.Infof("new connection %s", ip)
		tx.rxIP = ip

		for {
			select {
			case pkg, ok := <-ch:
				// mesage from tcp
				if !ok {
					logger.Errorf("tcp connection closed.")
					return nil
				}

				switch pkg.M.(type) {
				case *Signal:
					// signal from tcp
					signal := pkg.M.(*Signal)

					if signal.Action == 2 {
						// correct signal
						logger.Infof("signal from Rx, updating index...")

						if signal.Index == tx.rxIndex+1 {
							// latest image has been received
							tx.rxIndex = signal.Index

							newIndex := webserver.Index + 1
							atomic.StoreInt32(&webserver.Index, newIndex)

							logger.Infof("%d received by Rx", signal.Index)

						} else {
							// index error
							logger.Errorf("unexpected response from Rx: %s", signal)
						}

					} else {
						// unexpected signal
						logger.Errorf("invalid signal received: %s", signal)
						return nil
					}
				}
			}
		}
	}
}

func (tx *Transmitter) SendSignal() {
	signal := Signal{
		Index:  tx.txIndex,
		Action: 1,
	}
	tx.tcp.Send(signal, []string{tx.rxIP})
}

// func (tx *Transmitter) GetSendIndex() int64 {
// 	return tx.txIndex
// }
//
// func (tx *Transmitter) SendIndex(i int64) bool {
// 	logger.Infof("current Tx: %d, Rx: %d, Try to send: %d", tx.txIndex, tx.rxIndex, i)
//
// 	if tx.txIndex == 0 {
// 		logger.Infof("allow to send: 0")
// 		return true
// 	}
//
// 	if i == tx.rxIndex+1 {
// 		tx.txIndex = i
// 		logger.Infof("allow to send: %d", i)
// 		return true
// 	}
// 	logger.Warnf("%d is not allowed to send", i)
// 	return false
// }

func registerMessage(s *courier.Server) {
	s.Register((*Signal)(nil), 201)
}
