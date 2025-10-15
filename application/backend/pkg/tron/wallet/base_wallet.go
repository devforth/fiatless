package wallet

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fiatless/pkg/tron/address"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type BaseTronWallet struct {
	PrivateKey *ecdsa.PrivateKey
}

func (w *BaseTronWallet) GetAddress() *address.TronAddress {
	address, err := address.NewTronAddressFromPrivateKey(w.PrivateKey)
	if err != nil {
		return nil
	}
	return address
}

func NewBaseTronWallet(privateKey *ecdsa.PrivateKey) *BaseTronWallet {
	return &BaseTronWallet{
		PrivateKey: privateKey,
	}
}

func (w *BaseTronWallet) GetPrivateKey() *ecdsa.PrivateKey {
	return w.PrivateKey
}

func (w *BaseTronWallet) SignTransaction(msg []byte) ([]byte, error) {
	hash := sha256.Sum256(msg)

	signature, err := secp256k1.Sign(hash[:], crypto.FromECDSA(w.PrivateKey))
	if err != nil {
		return nil, err
	}

	return signature, nil
}
