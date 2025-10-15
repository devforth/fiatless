package solana

import (
	"fiatless/pkg/solana/address"

	"github.com/shopspring/decimal"
)

type CreateWalletRequestBody struct {
	TagID string `json:"tag_id,omitempty"`
}

type CreateWalletRequest struct {
	Payload *CreateWalletRequestBody `in:"body=json"`
}

type CreateWalletResponse struct {
	Address address.SolanaAddress `json:"address"`
	TagID   string                `json:"tag_id,omitempty"`
}

type GetBalanceRequest struct {
	Address string `in:"path=address"`
}

type GetBalanceResponse struct {
	Balance *decimal.Decimal `json:"sol_balance,omitempty"`
}

type WithdrawRequestBody struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type WithdrawRequest struct {
	Payload *WithdrawRequestBody `in:"body=json"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	ExplorerURL   string `json:"explorer_url"`
}
