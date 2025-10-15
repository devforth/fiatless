package solana

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/internal/models"
	"fiatless/pkg/solana"
	"log"
)

type BalanceHandler struct {
	Base   handler.BaseHandler
	Solana *solana.Solana
}

func NewBalanceHandler(client *solana.Solana) *BalanceHandler { return &BalanceHandler{Solana: client} }

func (h *BalanceHandler) CommandPath() string {
	return "/solana/wallet/balance"
}

func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.SolanaWalletBalanceRequest
	requestID, err := h.Base.ParseParams(command, &params)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	balanceLamports, err := h.Solana.GetBalance(ctx, params.Address)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("Solana balance retrieved: %v lamports", balanceLamports)

	result := models.SolanaWalletBalance{SOLBalance: h.Solana.LamportsToSOL(balanceLamports)}
	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": result,
	})
}
