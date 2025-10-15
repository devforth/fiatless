package wallet

import (
	"crypto/ecdsa"
	"fiatless/pkg/bsc/address"
)

type BaseBSCWallet struct {
	PrivateKey *ecdsa.PrivateKey
}

func (w *BaseBSCWallet) GetAddress() *address.BSCAddress {
	address, err := address.NewBSCAddressFromPrivateKey(w.PrivateKey)
	if err != nil {
		return nil
	}
	return address
}

func NewBaseBSCWallet(privateKey *ecdsa.PrivateKey) *BaseBSCWallet {
	return &BaseBSCWallet{
		PrivateKey: privateKey,
	}
}

func (w *BaseBSCWallet) GetPrivateKey() *ecdsa.PrivateKey {
	return w.PrivateKey
}
