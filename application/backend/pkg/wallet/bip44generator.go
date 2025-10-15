package wallet

import (
	"crypto/ecdsa"
	"fiatless/pkg/utils"

	"github.com/tyler-smith/go-bip32"
)

type BIP44KeyGenerator struct {
	mnemonic   string
	passphrase string
}

func NewBIP44KeyGenerator(mnemonic string, passphrase string) *BIP44KeyGenerator {
	return &BIP44KeyGenerator{
		mnemonic:   mnemonic,
		passphrase: passphrase,
	}
}

func (g *BIP44KeyGenerator) CreatePrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*ecdsa.PrivateKey, error) {
	bip32Key, err := utils.NewBIP44KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
	if err != nil {
		return nil, err
	}

	return utils.NewECDSAKeyFromBip32Key(bip32Key)
}

func (g *BIP44KeyGenerator) CreateBIP32Key(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*bip32.Key, error) {
	return utils.NewBIP44KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
}

func (g *BIP44KeyGenerator) GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string {
	return utils.GetBIP44DerivationPath(coinType, account, addressIndex)
}
