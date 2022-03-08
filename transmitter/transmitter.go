package transmitter

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	// "syncsampling/courier"

	"syncsampling/logs"
	"syncsampling/utils"
	"syncsampling/utils/tcpconn"
	"syncsampling/webserver"
)

var (
	tcp    tcpconn.ServerConfig
	logger *zap.SugaredLogger
)

type Transmitter struct {
	tcp     *tcpconn.Server
	tcpPort int // tcp host port, just for recorded

	rxIP string // IP of Rx
	// txIndex int64  // current index in Tx
	rxIndex int32 // current index in Rx

	actionCh chan int // receive action signal from webserver. 1-start
}

func NewTransmitter() (*Transmitter, error) {
	logger = logs.GetLogger()

	// host := "127.0.0.1"
	host := "0.0.0.0"
	port := 26001
	tcp = tcpconn.ServerConfig{
		Listen:      fmt.Sprintf("%s:%d", host, port),
		MaxConn:     10,
		MessageSize: 65535,
	}

	server, err := tcpconn.NewServer(tcp)
	if err != nil {
		return nil, err
	}

	m := Transmitter{
		tcp:     server,
		tcpPort: port,
		// txIndex:  0,
		rxIndex:  -1,
		actionCh: make(chan int, 0),
	}

	// reset webserver.Index to 0
	atomic.StoreInt32(&webserver.Index, 0)
	atomic.StoreInt32(&m.rxIndex, -1)
	webserver.InitActionCh(m.actionCh)

	return &m, nil
}

func (tx *Transmitter) Run() error {
	// start tcp server
	go func() {
		ip, _ := utils.GetOutBoundIP()
		logger.Infof("start tcp server at %s:%d", ip, tx.tcpPort)
		logger.Infof("                 or localhost:%d", tx.tcpPort)

		err := tx.tcp.Start()
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

	go tx.handleAction()

	return tx.handleTCP()
}

func (tx *Transmitter) handleAction() {
	for {
		a, ok := <-tx.actionCh
		if !ok {
			logger.Errorf("action chan close.")
			return
		}
		if a == 1 {
			logger.Infof("receive action signal")

			time.Sleep(time.Millisecond * 100)
			tx.SendSignal()
		}
	}
}

func (tx *Transmitter) handleTCP() error {
	for {
		// new tcp connection
		ip := <-tx.tcp.NewConnection()
		ch := tx.tcp.Receive(ip)
		logger.Infof("new connection %s", ip)
		tx.rxIP = ip

	OneConn:
		for {
			select {
			case pkg, ok := <-ch:
				// mesage from tcp
				if !ok {
					logger.Warnf("tcp connection closed.")
					atomic.StoreInt32(&webserver.Index, 0)
					atomic.StoreInt32(&tx.rxIndex, -1)
					break OneConn
				}
				// logger.Infof("receive raw from Rx: %s", pkg)

				signal := Signal{}
				err := json.Unmarshal(pkg, &signal)
				if err != nil {
					logger.Warnf("failed to decode: %s with raw: %s", err, pkg)
				}

				if signal.Action == 2 {
					// correct signal
					logger.Infof("signal from Rx, updating index...")

					if signal.Index == tx.rxIndex+1 {
						// latest image has been received
						tx.rxIndex = signal.Index

						logger.Infof("%d received by Rx", signal.Index)

						if webserver.Index < webserver.Total {
							newIndex := webserver.Index + 1
							atomic.StoreInt32(&webserver.Index, newIndex)
							logger.Infof("index update to %d\n", newIndex)
						}

					} else {
						// index error
						logger.Errorf("unexpected response from Rx with index %d, rxIndex is %d", signal, tx.rxIndex)
						tx.tcp.Disconnect(ip)
						break OneConn
					}

				} else {
					// unexpected signal
					logger.Errorf("invalid signal received: %s", signal)
					tx.tcp.Disconnect(ip)
					break OneConn
				}

			}
		}
	}
}

func (tx *Transmitter) SendSignal() {
	// Index :=  tx.txIndex,
	index := atomic.LoadInt32(&webserver.Index)

	signal := Signal{
		Index:  index,
		Action: 1,
	}
	logger.Infof("send message to Rx, index: %d", index)
	err := tx.tcp.Send(&signal, []string{tx.rxIP})
	if err != nil {
		logger.Errorf("failed to send to Rx: %s", err)
	}
}
