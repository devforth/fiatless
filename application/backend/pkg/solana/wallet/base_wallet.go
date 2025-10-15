package wallet

import (
	"crypto/ed25519"
	"fiatless/pkg/solana/address"
)

type BaseSolanaWallet struct {
	PrivateKey *ed25519.PrivateKey
}

func (w *BaseSolanaWallet) GetAddress() *address.SolanaAddress {
	address, err := address.NewSolanaAddressFromPrivateKey(*w.PrivateKey)
	if err != nil {
		return nil
	}
	return address
}

func NewBaseSolanaWallet(privateKey *ed25519.PrivateKey) *BaseSolanaWallet {
	return &BaseSolanaWallet{
		PrivateKey: privateKey,
	}
}

func (w *BaseSolanaWallet) GetPrivateKey() *ed25519.PrivateKey {
	return w.PrivateKey
}

func (w *BaseSolanaWallet) GetExpandedPrivateKey() []byte {
	return append(w.PrivateKey.Seed(), w.PrivateKey.Public().(ed25519.PublicKey)...)
}
