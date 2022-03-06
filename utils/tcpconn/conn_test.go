package tcpconn

import (
	"fmt"
	"testing"
	"time"
)

var (
	sConf = ServerConfig{
		Listen:      "127.0.0.1:46001",
		MaxConn:     10,
		MessageSize: 65535,
	}
	cConf = ClientConfig{
		Host:        "127.0.0.1:46001",
		MessageSize: 65535,
	}
)

func TestDisconnected(t *testing.T) {
	server, _ := NewServer(sConf)
	client, _ := NewClient(cConf)

	// start server
	go func() {
		err := server.Start()
		fmt.Printf("server exit: %s\n", err)
	}()

	time.Sleep(time.Second)

	go func() {
		for {
			select {
			case ip := <-server.NewConnection():
				fmt.Printf("[server] new connecion: %s\n", ip)
				go func() {
					ch := server.Receive(ip)
					if ch == nil {
						t.Fatalf("receive chan for %s is nil", ip)
					}
					for {
						m, ok := <-ch
						if !ok {
							fmt.Printf("%s conn closed\n", ip)
							break
						} else {
							fmt.Printf("server receive message:\n")
							fmt.Printf("    ip  : %s\n", ip)
							fmt.Printf("    data: %s\n", m)
						}
					}
				}()
			case s := <-server.Info():
				fmt.Printf("[INFO] [server]: %s\n", s)
			case e := <-server.Error():
				fmt.Printf("[ERROR] [server]: %s\n", e)
			case s := <-client.Info():
				fmt.Printf("[INFO] [client]: %s\n", s)
			case e := <-client.Error():
				fmt.Printf("[ERROR] [client]: %s\n", e)
			}
		}
	}()

	go func() {
		err := client.Connect()
		fmt.Printf("client exit: %v\n", err)
	}()

	time.Sleep(time.Second)
	err := client.Send("bad women")
	if err != nil {
		fmt.Println(err)
	}

	client.Disconnect()
	time.Sleep(time.Second)

	go func() {
		err := client.Connect()
		fmt.Printf("client exit: %v\n", err)
	}()

	time.Sleep(time.Second)
	err = client.Send("another bad women")
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 10)
}
