package bitcoin

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/bitcoin"
	"log"
)

type WithdrawHandler struct {
	handler.BitcoinHandler
}

func NewWithdrawHandler(bitcoin *bitcoin.Bitcoin) *WithdrawHandler {
	return &WithdrawHandler{
		BitcoinHandler: handler.NewBitcoinHandler(bitcoin),
	}
}

func (h *WithdrawHandler) CommandPath() string {
	return "/bitcoin/wallet/withdraw"
}

func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.BitcoinWithdrawRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	log.Printf("Withdrawing utxos: %v", params.UTXOs)
	var txID bp_models.BitcoinWithdrawResponse

	txID, err = h.Bitcoin.Withdraw(&params.ToAddress, params.Amount, params.UTXOs)
	if err != nil {
		return h.SendErrorResult(client, requestID, "Failed to withdraw BTC: "+err.Error())
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": txID,
	})
}
