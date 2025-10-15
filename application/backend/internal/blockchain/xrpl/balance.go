package xrpl

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/internal/models"

	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
)

type BalanceHandler struct {
	handler.XRPLHandler
}

func NewBalanceHandler(client *xrpl_rpc.Client) *BalanceHandler {
	return &BalanceHandler{
		XRPLHandler: handler.NewXRPLHandler(client),
	}
}

func (h *BalanceHandler) CommandPath() string {
	return "/xrpl/wallet/balance"
}

func (h *BalanceHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.XRPLWalletBalanceRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	balance, err := h.XRPL.GetXrpBalance(params.Address)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": models.XRPLWalletBalance{Balance: balance},
	})
}
