package wallet

import (
	"crypto/ecdsa"
	"crypto/ed25519"

	"github.com/tyler-smith/go-bip32"
)

type KeyGenerator interface {
	CreatePrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*ecdsa.PrivateKey, error)
	GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string
	CreateBIP32Key(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*bip32.Key, error)
}

// BIP86KeyGenerator derives keys using BIP-0086 (Taproot key-path addresses)
type BIP86KeyGenerator interface {
	KeyGenerator
}

// Ed25519KeyGenerator derives ed25519 private keys (for Solana)
type Ed25519KeyGenerator interface {
	CreateEd25519PrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (ed25519.PrivateKey, error)
	GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string
}
