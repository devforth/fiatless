package wallet

import (
	"crypto/ecdsa"
	"fiatless/pkg/ethereum/address"
)

type BaseEthereumWallet struct {
	PrivateKey *ecdsa.PrivateKey
}

func (w *BaseEthereumWallet) GetAddress() *address.EthereumAddress {
	address, err := address.NewEthereumAddressFromPrivateKey(w.PrivateKey)
	if err != nil {
		return nil
	}
	return address
}

func NewBaseEthereumWallet(privateKey *ecdsa.PrivateKey) *BaseEthereumWallet {
	return &BaseEthereumWallet{
		PrivateKey: privateKey,
	}
}

func (w *BaseEthereumWallet) GetPrivateKey() *ecdsa.PrivateKey {
	return w.PrivateKey
}
