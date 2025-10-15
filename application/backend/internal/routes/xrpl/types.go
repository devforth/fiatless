package xrpl

import (
	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

	"github.com/shopspring/decimal"
)

type CreateWalletRequestBody struct {
	TagID string `json:"tag_id,omitempty"`
}

type CreateWalletRequest struct {
	Payload *CreateWalletRequestBody `in:"body=json"`
}

type CreateWalletResponse struct {
	Address xrpl_types.Address `json:"address"`
	TagID   string             `json:"tag_id,omitempty"`
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
	Address string `in:"path=address"`
}

type GetBalanceResponse struct {
	Balance string `json:"xrp_balance"`
}
