package ethereum

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/ethereum/client"
	"log"
)

type BalanceHandler struct {
	handler.EthereumHandler
}

func NewBalanceHandler(client *client.EthereumClient) *BalanceHandler {
	return &BalanceHandler{
		EthereumHandler: handler.NewEthereumHandler(client),
	}
}

func (h *BalanceHandler) CommandPath() string {
	return "/ethereum/wallet/balance"
}

func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.EthereumWalletBalanceRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	balances, err := h.RpcClient.GetBalance(params.Address, params.IncludeETH, params.ERC20Tokens)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("Ethereum balances retrieved: %v", balances)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": balances,
	})
}
