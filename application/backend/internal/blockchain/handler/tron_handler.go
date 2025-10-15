package handler

import (
	"fiatless/pkg/tron"
)

type TronHandler struct {
	BaseHandler
	Tron *tron.Tron
}

func NewTronHandler(tron *tron.Tron) TronHandler {
	return TronHandler{
		Tron: tron,
	}
}
