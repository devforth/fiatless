package wallet

import (
	"crypto/ecdsa"
	"fiatless/pkg/utils"

	"github.com/tyler-smith/go-bip32"
)

type BIP84KeyGenerator struct {
	mnemonic   string
	passphrase string
}

func NewBIP84KeyGenerator(mnemonic string, passphrase string) *BIP84KeyGenerator {
	return &BIP84KeyGenerator{
		mnemonic:   mnemonic,
		passphrase: passphrase,
	}
}

func (g *BIP84KeyGenerator) CreatePrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*ecdsa.PrivateKey, error) {
	bip32Key, err := utils.NewBIP84KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
	if err != nil {
		return nil, err
	}

	return utils.NewECDSAKeyFromBip32Key(bip32Key)
}

func (g *BIP84KeyGenerator) CreateBIP32Key(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*bip32.Key, error) {
	return utils.NewBIP84KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
}

func (g *BIP84KeyGenerator) GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string {
	return utils.GetBIP84DerivationPath(coinType, account, addressIndex)
}
