package wallet

import (
	"crypto/ecdsa"
	"fiatless/pkg/utils"

	bip32 "github.com/tyler-smith/go-bip32"
)

type BIP86KeyGen struct {
	mnemonic   string
	passphrase string
}

func NewBIP86KeyGenerator(mnemonic string, passphrase string) *BIP86KeyGen {
	return &BIP86KeyGen{mnemonic: mnemonic, passphrase: passphrase}
}

func (g *BIP86KeyGen) CreatePrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*ecdsa.PrivateKey, error) {
	bip32Key, err := utils.NewBIP86KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
	if err != nil {
		return nil, err
	}
	return utils.NewECDSAKeyFromBip32Key(bip32Key)
}

func (g *BIP86KeyGen) CreateBIP32Key(coinType uint32, account uint32, chain uint32, addressIndex uint32) (*bip32.Key, error) {
	return utils.NewBIP86KeyFromMnemonic(g.mnemonic, g.passphrase, coinType, account, chain, addressIndex)
}

func (g *BIP86KeyGen) GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string {
	return utils.GetBIP86DerivationPath(coinType, account, addressIndex)
}
