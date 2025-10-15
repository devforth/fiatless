package wallet

import (
	"crypto/ecdsa"
	"fiatless/internal/models"
	"fiatless/internal/services/bitcoin"
	"fiatless/pkg/bitcoin/address"

	"github.com/btcsuite/btcd/btcec/v2"
)

type BitcoinWallet struct {
	BaseBitcoinWallet
	bitcoinService *bitcoin.Service
}

func NewBitcoinWallet(privateKey *ecdsa.PrivateKey, bitcoinService *bitcoin.Service, network models.BitcoinNetwork, addressType address.BitcoinAddressType) *BitcoinWallet {
	btcecPrivateKey, _ := btcec.PrivKeyFromBytes(privateKey.D.Bytes())
	return &BitcoinWallet{
		BaseBitcoinWallet: BaseBitcoinWallet{
			PrivateKey:  btcecPrivateKey,
			AddressType: addressType,
			Network:     network,
		},
		bitcoinService: bitcoinService,
	}
}

func (w *BitcoinWallet) GetBalance() (*models.BitcoinWalletBalance, error) {
	balance, err := w.bitcoinService.GetBalance(*w.GetAddress())
	if err != nil {
		return nil, err
	}
	return balance, nil
}

// func (w *BitcoinWallet) Withdraw(toAddress *address.BitcoinAddress, amount decimal.Decimal, utxos []models.UTXO) (models.BitcoinWithdrawResponse, error) {
// 	withdrawResponse, err := w.bitcoinService.Withdraw(w.GetPrivateKey(), toAddress, amount, utxos)
// 	if err != nil {
// 		return models.BitcoinWithdrawResponse{}, err
// 	}
// 	return withdrawResponse, nil
// }
