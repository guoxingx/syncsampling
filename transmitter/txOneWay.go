package transmitter

import (
	"fmt"
	"sync/atomic"
	"time"

	// "syncsampling/courier"

	"syncsampling/logs"
	"syncsampling/utils"
	"syncsampling/utils/tcpconn"
	"syncsampling/webserver"
)

type Txow struct {
	tcp     *tcpconn.Server
	tcpPort int // tcp host port, just for recorded

	rxIP       string   // IP of Rx
	actionCh   chan int // receive action signal from webserver. 1-start
	interval   time.Duration
	lastSignal time.Time
}

func NewTxow() (*Txow, error) {
	logger = logs.GetLogger()

	// host := "127.0.0.1"
	host := "0.0.0.0"
	port := 26001
	interval := time.Second * 3

	tcp = tcpconn.ServerConfig{
		Listen:      fmt.Sprintf("%s:%d", host, port),
		MaxConn:     10,
		MessageSize: 65535,
	}

	server, err := tcpconn.NewServer(tcp)
	if err != nil {
		return nil, err
	}

	m := Txow{
		tcp:      server,
		tcpPort:  port,
		actionCh: make(chan int, 0),
		interval: interval,
	}

	// reset webserver.Index to 0
	atomic.StoreInt32(&webserver.Index, 0)
	webserver.InitActionCh(m.actionCh)

	return &m, nil
}

func (tx *Txow) Run() error {
	// start tcp server
	go func() {
		ip, _ := utils.GetOutBoundIP()
		logger.Infof("start tcp server at %s:%d", ip, tx.tcpPort)

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

func (tx *Txow) handleAction() {
	for {
		a, ok := <-tx.actionCh
		if !ok {
			logger.Errorf("action chan close.")
			return
		}
		if a == 1 {
			logger.Infof("receive action signal")

			time.Sleep(time.Millisecond * 100)

			// Index :=  tx.txIndex,
			index := atomic.LoadInt32(&webserver.Index)

			interval := time.Second * 30
			logger.Infof("signal for index %d will sent in %s", index, interval)
			go tx.SendSignal(index, interval)
			go simuRx(tx.interval)
		}
	}
}

func (tx *Txow) handleTCP() error {
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
					break OneConn
				}
				logger.Infof("%s received from client, ignore.", pkg)
			}
		}
	}
}

func (tx *Txow) SendSignal(index int32, d time.Duration) {
	// Unix		1646720683
	// UnixNano 1646720683479108000
	// required 1646720683479
	ts := time.Now().UnixNano()
	ts = ts / 1000000

	if err := utils.WriteLineToFile("signals_tx.txt", fmt.Sprintf("%d", ts)); err != nil {
		logger.Warnf("failed to backup signal: %v", err)
	}

	time.Sleep(d)

	signal := Signal{
		Index:  index,
		Action: 4,
		Time:   ts,
	}
	logger.Infof("send message to Rx, index: %d", index)
	err := tx.tcp.Send(&signal, []string{tx.rxIP})
	if err != nil {
		logger.Errorf("failed to send to Rx: %s", err)
	}

}

func simuRx(d time.Duration) {
	time.Sleep(d)
	newIndex := atomic.LoadInt32(&webserver.Index) + 1
	atomic.StoreInt32(&webserver.Index, newIndex)
	logger.Infof("image with index %d is allowd to show", newIndex)
}
