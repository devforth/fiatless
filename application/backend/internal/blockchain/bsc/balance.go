package bsc

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/bsc/client"
	"log"
)

// BalanceHandler handles Ethereum balance requests
type BalanceHandler struct {
	handler.BSCHandler
}

// NewBalanceHandler creates a new Ethereum balance handler
func NewBalanceHandler(client *client.BSCClient) *BalanceHandler {
	return &BalanceHandler{
		BSCHandler: handler.NewBSCHandler(client),
	}
}

// CommandPath returns the command path for this handler
func (h *BalanceHandler) CommandPath() string {
	return "/bsc/wallet/balance"
}

// Handle processes the balance request
func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.BSCWalletBalanceRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	balances, err := h.RpcClient.GetBalance(params.Address, params.IncludeBNB, params.BEP20Tokens)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("BSC balances retrieved: %v", balances)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": balances,
	})
}
