package webserver

type Config struct {
	Interval int `json:"interval"`
}

type Action uint8

const (
	_ Action = iota
	ActionStart
	ActionReady
	ActionStop
	ActionPause
	ActionContinue
)
