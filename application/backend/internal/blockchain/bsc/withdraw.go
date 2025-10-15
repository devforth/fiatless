package bsc

import (
	"context"
	"fiatless/internal/blockchain/handler"
	"fiatless/internal/ijson"
	"fiatless/internal/models"
	"fiatless/pkg/bsc/client"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
)

// WithdrawHandler handles Ethereum withdraw requests
type WithdrawHandler struct {
	handler.BSCHandler
}

// NewWithdrawHandler creates a new Ethereum withdraw handler
func NewWithdrawHandler(client *client.BSCClient) *WithdrawHandler {
	return &WithdrawHandler{
		BSCHandler: handler.NewBSCHandler(client),
	}
}

// CommandPath returns the command path for this handler
func (h *WithdrawHandler) CommandPath() string {
	return "/bsc/wallet/withdraw"
}

// Handle processes the withdraw request
func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params models.BSCWithdrawRequest
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

	tx, err := h.RpcClient.Withdraw(context.Background(), privKey, &params.ToAddress, &params.Amount, params.ContractAddress)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	log.Printf("BSC transaction completed: %v", tx)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": tx,
	})
}
