package xrpl

import xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"

type XRPL struct {
	client *xrpl_rpc.Client
}

func NewXRPL(client *xrpl_rpc.Client) *XRPL {
	return &XRPL{client: client}
}
