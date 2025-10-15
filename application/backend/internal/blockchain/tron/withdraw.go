package tron

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/tron"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
)

type WithdrawHandler struct {
	handler.TronHandler
}

func NewWithdrawHandler(tron *tron.Tron) *WithdrawHandler {
	return &WithdrawHandler{
		TronHandler: handler.NewTronHandler(tron),
	}
}

func (h *WithdrawHandler) CommandPath() string {
	return "/tron/wallet/withdraw"
}

func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.TronWithdrawRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	privBytes, err := base58.Decode(params.PrivateKey)
	if err != nil {
		return h.SendErrorResult(client, requestID, "Invalid private key format")
	}
	privKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return h.SendErrorResult(client, requestID, "Invalid private key format")
	}

	var txID bp_models.TronWithdrawResponse

	if params.ContractAddress != nil {
		txID, err = h.Tron.WithdrawTRC20(privKey, &params.ToAddress, params.Amount, params.ContractAddress)
		if err != nil {
			return h.SendErrorResult(client, requestID, "Failed to withdraw TRC20: "+err.Error())
		}
	} else {
		txID, err = h.Tron.WithdrawTRX(privKey, &params.ToAddress, params.Amount)
		if err != nil {
			return h.SendErrorResult(client, requestID, "Failed to withdraw TRX: "+err.Error())
		}
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": txID,
	})
}
