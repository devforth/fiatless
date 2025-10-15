package wallet

import "crypto/ecdsa"

type Wallet interface {
	GetPrivateKey() *ecdsa.PrivateKey

	GetAddress() WalletAddress

	Sign(message []byte) ([]byte, error)
}

type WalletAddress interface {
	GetAddress() string
}
