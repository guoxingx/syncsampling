package transmitter

// Action
//
// 1: sent, from Tx to Rx
// 2: received, from Rx to Tx
// 3: falied to receive, from Rx to Tx
// 4: one way signal
type Signal struct {
	Index  int32 `json:"i"`
	Action uint8 `json:"a"`
	Time   int64 `json:"t"`
}
