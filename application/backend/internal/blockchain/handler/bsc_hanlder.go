package handler

import "fiatless/pkg/bsc/client"

type BSCHandler struct {
	BaseHandler
	RpcClient *client.BSCClient
}

func NewBSCHandler(client *client.BSCClient) BSCHandler {
	return BSCHandler{
		RpcClient: client,
	}
}
