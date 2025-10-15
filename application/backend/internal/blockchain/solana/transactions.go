package solana

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/constants"
	"fiatless/internal/ijson"
	"fiatless/pkg/solana"

	"github.com/shopspring/decimal"
)

// WithdrawHandler handles Solana withdraw requests
type TransactionsHandler struct {
	Base   handler.BaseHandler
	Solana *solana.Solana
}

func NewTransactionsHandler(client *solana.Solana) *TransactionsHandler {
	return &TransactionsHandler{Solana: client}
}

func (h *TransactionsHandler) CommandPath() string {
	return "/solana/wallet/transactions"
}

// Handle processes the withdraw request
func (h *TransactionsHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.SolanaGetWalletTransactionsRequest
	requestID, err := h.Base.ParseParams(command, &params)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	signatures, err := h.Solana.GetWalletSignatures(ctx, params.WalletAddress, params.LatestTransactionId)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	transactions, err := h.Solana.GetTransactions(ctx, signatures)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	result := bp_models.SolanaGetWalletTransactionsResponse{
		Transactions: make([]bp_models.SolanaTransaction, len(transactions)),
	}
	for i, transaction := range transactions {
		tx, err := transaction.Transaction.GetTransaction()
		if err != nil {
			return h.Base.SendErrorResult(client, requestID, err.Error())
		}
		for y, key := range tx.Message.AccountKeys {
			if key.String() == params.WalletAddress.String() {
				result.Transactions[i] = bp_models.SolanaTransaction{
					TxID:      tx.Signatures[0].String(),
					Address:   key.String(),
					Fee:       decimal.NewFromInt(int64(transaction.Meta.Fee)).Shift(-int32(constants.SolanaDecimals)),
					Amount:    decimal.NewFromInt(int64(transaction.Meta.PostBalances[y] - transaction.Meta.PreBalances[y])).Shift(-int32(constants.SolanaDecimals)),
					Timestamp: int64(*transaction.BlockTime),
				}
				break
			}
		}
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": result,
	})
}
