package utils

import (
	"fmt"

	bip32 "github.com/tyler-smith/go-bip32"
)

const bip84Purpose uint32 = 0x80000054

func GetBIP84DerivationPath(coinType uint32, account, index uint32) string {
	return fmt.Sprintf("m/84'/%d'/%d'/0/%d", coinType, account, index)
}

func NewBIP84KeyFromMnemonic(mnemonic string, passphrase string, coin uint32, account uint32, chain uint32, address_index uint32) (*bip32.Key, error) {
	return NewKeyFromMnemonic(mnemonic, passphrase, bip84Purpose, coin, account, chain, address_index)
}
