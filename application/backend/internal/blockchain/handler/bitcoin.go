package handler

import (
	"fiatless/pkg/bitcoin"
)

type BitcoinHandler struct {
	BaseHandler
	Bitcoin *bitcoin.Bitcoin
}

func NewBitcoinHandler(client *bitcoin.Bitcoin) BitcoinHandler {
	return BitcoinHandler{
		Bitcoin: client,
	}
}
