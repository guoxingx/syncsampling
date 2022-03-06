package transmitter

// Action
//
// 1: sent, from Tx to Rx
// 2: received, from Rx to Tx
// 3: falied to receive, from Rx to Tx
type Signal struct {
	Index  int64  `json:"i"`
	Action uint8  `json:"a"`
	ErrS   string `json:"e"`
}
