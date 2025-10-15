package utils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
)

func NewKeyFromMasterKey(masterKey *bip32.Key, purpose uint32, coin uint32, account uint32, chain uint32, address_index uint32) (*bip32.Key, error) {
	child, err := masterKey.NewChildKey(purpose)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(bip32.FirstHardenedChild + coin)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(bip32.FirstHardenedChild + account)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(chain)
	if err != nil {
		return nil, err
	}

	child, err = child.NewChildKey(address_index)
	if err != nil {
		return nil, err
	}

	return child, nil
}

func NewKeyFromMnemonic(mnemonic string, passphrase string, purpose uint32, coin uint32, account uint32, chain uint32, address_index uint32) (*bip32.Key, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return NewKeyFromMasterKey(masterKey, purpose, coin, account, chain, address_index)
}

func NewECDSAKeyFromBip32Key(bip32Key *bip32.Key) (*ecdsa.PrivateKey, error) {
	return crypto.ToECDSA(bip32Key.Key)
}
