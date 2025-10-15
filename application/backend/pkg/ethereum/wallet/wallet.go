package wallet

import (
	"crypto/ecdsa"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/internal/services/ethereum"
	"fiatless/pkg/ethereum/address"

	"github.com/shopspring/decimal"
)

type EthereumWallet struct {
	BaseEthereumWallet
	ethereumService *ethereum.Service
}

func NewEthereumWallet(privateKey *ecdsa.PrivateKey, ethereumService *ethereum.Service) *EthereumWallet {
	return &EthereumWallet{
		BaseEthereumWallet: BaseEthereumWallet{
			PrivateKey: privateKey,
		},
		ethereumService: ethereumService,
	}
}

func (w *EthereumWallet) GetBalance(includeETH bool, erc20Tokens []address.EthereumAddress) (*models.EthereumWalletBalance, error) {
	balance, err := w.ethereumService.GetBalance(*w.GetAddress(), includeETH, erc20Tokens)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *EthereumWallet) Withdraw(toAddress *address.EthereumAddress, amount decimal.Decimal, tokenAddress *address.EthereumAddress) (bp_models.EthereumWithdrawResponse, error) {
	withdrawResponse, err := w.ethereumService.Withdraw(w.GetPrivateKey(), toAddress, amount, tokenAddress)
	if err != nil {
		return bp_models.EthereumWithdrawResponse{}, err
	}
	return withdrawResponse, nil
}
