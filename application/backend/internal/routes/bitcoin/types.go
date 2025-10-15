package bitcoin

import (
	"fiatless/pkg/bitcoin/address"

	"github.com/shopspring/decimal"
)

type GetBalanceRequest struct {
	Address string `in:"path=address"`
}

type GetBalanceResponse struct {
	Balance *decimal.Decimal `json:"btc_balance,omitempty"`
}

type CreateWalletRequestBody struct {
	TagID string `json:"tag_id,omitempty"`
}

type CreateWalletRequest struct {
	Payload *CreateWalletRequestBody `in:"body=json"`
}

type CreateWalletResponse struct {
	Address address.BitcoinAddress `json:"address"`
	TagID   string                 `json:"tag_id,omitempty"`
}

type WithdrawRequestBody struct {
	Address address.BitcoinAddress `json:"address"`
	Amount  decimal.Decimal        `json:"amount"`
}

type WithdrawRequest struct {
	Payload *WithdrawRequestBody `in:"body=json"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	ExplorerURL   string `json:"explorer_url"`
}
