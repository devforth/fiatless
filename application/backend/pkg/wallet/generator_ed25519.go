package wallet

import (
	"crypto/ed25519"
	"fmt"

	"fiatless/pkg/utils"

	bip39 "github.com/tyler-smith/go-bip39"
)

type SLIP10Ed25519KeyGenerator struct {
	mnemonic   string
	passphrase string
}

func NewSLIP10Ed25519KeyGenerator(mnemonic string, passphrase string) *SLIP10Ed25519KeyGenerator {
	return &SLIP10Ed25519KeyGenerator{mnemonic: mnemonic, passphrase: passphrase}
}

func (g *SLIP10Ed25519KeyGenerator) seed() ([]byte, error) {
	if !bip39.IsMnemonicValid(g.mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	return bip39.NewSeed(g.mnemonic, g.passphrase), nil
}

func (g *SLIP10Ed25519KeyGenerator) CreateEd25519PrivateKey(coinType uint32, account uint32, chain uint32, addressIndex uint32) (ed25519.PrivateKey, error) {
	seed, err := g.seed()
	if err != nil {
		return nil, err
	}
	master, err := utils.NewSlip10Master(seed)
	if err != nil {
		return nil, err
	}
	node, err := master.DerivePath(
		utils.SolanaPurpose(),
		utils.SolanaCoinTypeHardened(),
		utils.SolanaAccount(account),
		utils.SolanaChange(chain),
		utils.SolanaAddressIndex(addressIndex),
	)
	if err != nil {
		return nil, err
	}
	return node.PrivateKeyEd25519(), nil
}

func (g *SLIP10Ed25519KeyGenerator) GetDerivationPath(coinType uint32, account uint32, chain uint32, addressIndex uint32) string {
	return fmt.Sprintf("m/44'/%d'/%d'/%d'/%d'", coinType, account, chain, addressIndex)
}
