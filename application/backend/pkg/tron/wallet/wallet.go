package wallet

import (
	"context"
	"crypto/ecdsa"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"fiatless/internal/services/tron"
	"fiatless/pkg/tron/address"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TronWallet struct {
	BaseTronWallet
	tronService *tron.Service
}

func NewTronWallet(privateKey *ecdsa.PrivateKey, tronService *tron.Service) *TronWallet {
	return &TronWallet{
		BaseTronWallet: BaseTronWallet{
			PrivateKey: privateKey,
		},
		tronService: tronService,
	}
}

func (w *TronWallet) GetBalance(includeTRX bool, includeResources bool, includeTRC20 []address.TronAddress, includeTRC10 []string) (*models.TronWalletBalance, error) {
	balance, err := w.tronService.GetBalance(*w.GetAddress(), includeTRX, includeResources, includeTRC20, includeTRC10)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *TronWallet) GetBalanceOffChain(walletRepository repositories.WalletRepository, tokenID uuid.UUID) (*decimal.Decimal, error) {
	balance, err := walletRepository.Balance(context.Background(), w.GetAddress().String(), tokenID)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *TronWallet) Withdraw(toAddress *address.TronAddress, amount decimal.Decimal, tokenAddress *address.TronAddress) (bp_models.TronWithdrawResponse, error) {
	withdrawResponse, err := w.tronService.Withdraw(w.GetPrivateKey(), *toAddress, amount, tokenAddress)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, err
	}
	return withdrawResponse, nil
}
