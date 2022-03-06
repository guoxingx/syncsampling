package tcpconn

// ServerConfig ...
type ServerConfig struct {
	Listen      string `json:"listen"`
	MaxConn     int    `json:"max_conn"`
	MessageSize int    `json:"message_size"`
}

// ClientConfig ...
type ClientConfig struct {
	Host        string `json:"host"`
	MessageSize int    `json:"message_size"`
}

// Message ...
type Message struct {
	IP   string
	Data []byte
}

type sendReq struct {
	v   interface{}
	ips []string
}
