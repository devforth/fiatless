package bitcoin

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/bitcoin"
	"log"
)

// BalanceHandler handles Ethereum balance requests
type BalanceHandler struct {
	handler.BitcoinHandler
}

// NewBalanceHandler creates a new Ethereum balance handler
func NewBalanceHandler(client *bitcoin.Bitcoin) *BalanceHandler {
	return &BalanceHandler{
		BitcoinHandler: handler.NewBitcoinHandler(client),
	}
}

// CommandPath returns the command path for this handler
func (h *BalanceHandler) CommandPath() string {
	return "/bitcoin/wallet/balance"
}

// Handle processes the balance request
func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.BitcoinWalletBalanceRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	balances, err := h.Bitcoin.GetBalance(params.Address)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("Bitcoin balances retrieved: %v", balances)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": balances.String(),
	})
}
