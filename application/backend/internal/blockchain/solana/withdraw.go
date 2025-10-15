package solana

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/solana"
	"log"
)

// WithdrawHandler handles Solana withdraw requests
type WithdrawHandler struct {
	Base   handler.BaseHandler
	Solana *solana.Solana
}

func NewWithdrawHandler(client *solana.Solana) *WithdrawHandler {
	return &WithdrawHandler{Solana: client}
}

func (h *WithdrawHandler) CommandPath() string {
	return "/solana/wallet/withdraw"
}

// Handle processes the withdraw request
func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.SolanaWithdrawRequest
	requestID, err := h.Base.ParseParams(command, &params)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	tx, err := h.Solana.TransferSOL(ctx, params.ExpandedPrivateKey, params.ToAddress, params.Amount)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("Solana transaction completed: %v", tx)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": tx,
	})
}
