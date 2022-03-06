package main

import (
	"syncsampling/logs"
	"syncsampling/transmitter"
)

func main() {
	logger := logs.GetLogger()

	tx, _ := transmitter.NewTransmitter()
	err := tx.Run()
	logger.Fatalf("Tx exit: %s", err)
}
