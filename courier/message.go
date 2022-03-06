package courier

// Message interface
// type Message interface {
// 	Encode() ([]byte, error)
//
// 	Decode([]byte) (Message, error)
// }

// Package wraps Message and other informations
type Package struct {
	// M     Message     `json:"m"`
	M     interface{} `json:"m"`
	From  string      `json:"f,omitempty"`
	Error error       `json:"e,omitempty"`
}
