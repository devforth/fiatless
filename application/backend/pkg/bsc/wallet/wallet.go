package wallet

import (
	"crypto/ecdsa"
	"fiatless/internal/models"
	"fiatless/internal/services/bsc"
	"fiatless/pkg/bsc/address"

	"github.com/shopspring/decimal"
)

type BSCWallet struct {
	BaseBSCWallet
	bscService *bsc.Service
}

func NewBSCWallet(privateKey *ecdsa.PrivateKey, bscService *bsc.Service) *BSCWallet {
	return &BSCWallet{
		BaseBSCWallet: BaseBSCWallet{
			PrivateKey: privateKey,
		},
		bscService: bscService,
	}
}

func (w *BSCWallet) GetBalance(includeBNB bool, bep20Tokens []address.BSCAddress) (*models.BSCWalletBalance, error) {
	balance, err := w.bscService.GetBalance(*w.GetAddress(), includeBNB, bep20Tokens)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *BSCWallet) Withdraw(toAddress *address.BSCAddress, amount decimal.Decimal, tokenAddress *address.BSCAddress) (models.BSCWithdrawResponse, error) {
	withdrawResponse, err := w.bscService.Withdraw(w.GetPrivateKey(), toAddress, amount, tokenAddress)
	if err != nil {
		return models.BSCWithdrawResponse{}, err
	}
	return withdrawResponse, nil
}
