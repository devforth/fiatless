package wallet

import (
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/internal/services/xrpl"

	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/shopspring/decimal"
	"github.com/tyler-smith/go-bip32"
)

type XRPLWallet struct {
	BaseXRPLWallet
	xrplService *xrpl.Service
}

func NewXRPLWallet(bip32Key *bip32.Key, xrplService *xrpl.Service) *XRPLWallet {
	return &XRPLWallet{BaseXRPLWallet: BaseXRPLWallet{wallet: FromBip32Key(bip32Key)}, xrplService: xrplService}
}

func (w *XRPLWallet) GetBalance() (*models.XRPLWalletBalance, error) {
	balance, err := w.xrplService.GetBalance(w.GetAddress())
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (w *XRPLWallet) Withdraw(toAddress xrpl_types.Address, amount decimal.Decimal) (bp_models.XRPLWithdrawResponse, error) {
	withdrawResponse, err := w.xrplService.Withdraw(w.GetPrivateKey(), toAddress, amount)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, err
	}
	return withdrawResponse, nil
}
