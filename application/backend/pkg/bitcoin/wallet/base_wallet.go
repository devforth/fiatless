package wallet

import (
	"crypto/ecdsa"
	"fiatless/internal/models"
	"fiatless/pkg/bitcoin/address"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
)

type BaseBitcoinWallet struct {
	PrivateKey  *btcec.PrivateKey
	AddressType address.BitcoinAddressType
	Network     models.BitcoinNetwork
}

func (w *BaseBitcoinWallet) GetAddress() *address.BitcoinAddress {
	addr, err := address.NewBitcoinAddressFromPrivateKey(w.PrivateKey, w.AddressType, w.Network)
	log.Println("err", err)
	if err != nil {
		return nil
	}
	return addr
}

func NewBaseBitcoinWallet(privateKey *ecdsa.PrivateKey, addressType address.BitcoinAddressType, network models.BitcoinNetwork) *BaseBitcoinWallet {
	btcecPrivateKey, _ := btcec.PrivKeyFromBytes(privateKey.D.Bytes())
	return &BaseBitcoinWallet{
		PrivateKey:  btcecPrivateKey,
		AddressType: addressType,
		Network:     network,
	}
}

func (w *BaseBitcoinWallet) GetPrivateKey() *btcec.PrivateKey {
	return w.PrivateKey
}

func (w *BaseBitcoinWallet) GetAddressType() address.BitcoinAddressType {
	return w.AddressType
}

func (w *BaseBitcoinWallet) GetNetwork() models.BitcoinNetwork {
	return w.Network
}
