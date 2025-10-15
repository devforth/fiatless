package ethereum

import (
	"fiatless/pkg/ethereum/address"

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
	Address address.EthereumAddress `json:"address"`
	TagID   string                  `json:"tag_id,omitempty"`
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
	IncludeERC20 bool   `in:"query=include_erc20"`
	IncludeETH   bool   `in:"query=include_eth"`
}

type GetMainWalletBalanceRequest struct {
	IncludeERC20 bool `in:"query=include_erc20"`
	IncludeETH   bool `in:"query=include_eth"`
}

type ERC20Balance struct {
	TokenID uuid.UUID       `json:"token_id"`
	Balance decimal.Decimal `json:"balance"`
}

type GetBalanceResponse struct {
	Balance       *decimal.Decimal `json:"eth_balance,omitempty"`
	ERC20Balances []ERC20Balance   `json:"erc20_balances,omitempty"`
}
