package wallet

import (
	"crypto/ed25519"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/internal/services/solana"
	"fiatless/pkg/solana/address"

	"github.com/shopspring/decimal"
)

type SolanaWallet struct {
	BaseSolanaWallet
	solanaService *solana.Service
}

func NewSolanaWallet(privateKey *ed25519.PrivateKey, solanaService *solana.Service) *SolanaWallet {
	return &SolanaWallet{
		BaseSolanaWallet: BaseSolanaWallet{
			PrivateKey: privateKey,
		},
		solanaService: solanaService,
	}
}

func (w *SolanaWallet) GetBalance() (*models.SolanaWalletBalance, error) {
	balance, err := w.solanaService.GetBalance(*w.GetAddress())
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *SolanaWallet) Withdraw(toAddress *address.SolanaAddress, amount decimal.Decimal) (bp_models.SolanaWithdrawResponse, error) {
	withdrawResponse, err := w.solanaService.Withdraw(w.GetExpandedPrivateKey(), *toAddress, amount)
	if err != nil {
		return bp_models.SolanaWithdrawResponse{}, err
	}
	return withdrawResponse, nil
}
