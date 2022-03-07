package main

import (
	"syncsampling/transmitter"
)

func main() {
	// tx, _ := transmitter.NewTransmitter()
	tx, _ := transmitter.NewTxow()
	tx.Run()
}
