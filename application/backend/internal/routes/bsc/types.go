package bsc

import (
	"fiatless/pkg/bsc/address"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateWalletRequestBody struct {
	TagID string `json:"tag_id,omitempty"`
}

type CreateWalletRequest struct {
	Payload *CreateWalletRequestBody `in:"body=json"`
}

type CreateWalletResponse struct {
	Address address.BSCAddress `json:"address"`
	TagID   string             `json:"tag_id,omitempty"`
}

type WithdrawRequestBody struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
	TokenID string `json:"token_id,omitempty"`
}

type WithdrawRequest struct {
	Payload *WithdrawRequestBody `in:"body=json"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	ExplorerURL   string `json:"explorer_url"`
}

type GetBalanceRequest struct {
	Address      string `in:"path=address"`
	IncludeBNB   bool   `in:"query=include_bnb"`
	IncludeBEP20 bool   `in:"query=include_bep20"`
}

type GetMainWalletBalanceRequest struct {
	IncludeBNB   bool `in:"query=include_bnb"`
	IncludeBEP20 bool `in:"query=include_bep20"`
}

type BEP20Balance struct {
	TokenID uuid.UUID       `json:"token_id"`
	Balance decimal.Decimal `json:"balance"`
}

type GetBalanceResponse struct {
	Balance       *decimal.Decimal `json:"bnb_balance,omitempty"`
	BEP20Balances []BEP20Balance   `json:"bep20_balances,omitempty"`
}
