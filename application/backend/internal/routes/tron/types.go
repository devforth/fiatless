package tron

import (
	"fiatless/pkg/tron/address"

	"github.com/shopspring/decimal"
)

type CreateWalletRequestBody struct {
	TagID string `json:"tag_id,omitempty"`
}

type CreateWalletRequest struct {
	Payload *CreateWalletRequestBody `in:"body=json"`
}

type CreateWalletResponse struct {
	Address address.TronAddress `json:"address"`
	TagID   string              `json:"tag_id,omitempty"`
}

type WithdrawRequestBody struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
	Token   string `json:"token,omitempty"`
}

type WithdrawRequest struct {
	Payload *WithdrawRequestBody `in:"body=json"`
}

type StartSweepingRequestBody struct {
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`
	TokenID   string           `json:"token_id"`
}

type StartSweepingRequest struct {
	Payload *StartSweepingRequestBody `in:"body=json"`
}

type StartSweepingResponse struct {
	SessionID string `json:"session_id"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	ExplorerURL   string `json:"explorer_url"`
}

type GetBalanceRequest struct {
	Address          string `in:"path=address"`
	IncludeTRX       bool   `in:"query=include_trx"`
	IncludeResources bool   `in:"query=include_resources"`
	IncludeTRC20     bool   `in:"query=include_trc20"`
	IncludeTRC10     bool   `in:"query=include_trc10"`
}

type GetMainWalletBalanceRequest struct {
	IncludeTRX       bool `in:"query=include_trx"`
	IncludeResources bool `in:"query=include_resources"`
	IncludeTRC20     bool `in:"query=include_trc20"`
	IncludeTRC10     bool `in:"query=include_trc10"`
}
