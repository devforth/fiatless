package ethereum

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/ethereum/client"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
)

// WithdrawHandler handles Ethereum withdraw requests
type WithdrawHandler struct {
	handler.EthereumHandler
}

// NewWithdrawHandler creates a new Ethereum withdraw handler
func NewWithdrawHandler(client *client.EthereumClient) *WithdrawHandler {
	return &WithdrawHandler{
		EthereumHandler: handler.NewEthereumHandler(client),
	}
}

// CommandPath returns the command path for this handler
func (h *WithdrawHandler) CommandPath() string {
	return "/ethereum/wallet/withdraw"
}

// Handle processes the withdraw request
func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.EthereumWithdrawRequest
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

	log.Printf("Ethereum transaction completed: %v", tx)

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": tx,
	})
}
