package handler

import (
	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
)

type XRPLHandler struct {
	BaseHandler
	XRPL *xrpl_rpc.Client
}

func NewXRPLHandler(xrpl *xrpl_rpc.Client) XRPLHandler {
	return XRPLHandler{
		XRPL: xrpl,
	}
}
