package handler

import "fiatless/pkg/ethereum/client"

type EthereumHandler struct {
	BaseHandler
	RpcClient *client.EthereumClient
}

func NewEthereumHandler(client *client.EthereumClient) EthereumHandler {
	return EthereumHandler{
		RpcClient: client,
	}
}
