package wallet

import "crypto/ecdsa"

// WalletManager defines the common interface for all wallet managers
type WalletManager interface {
	// CreateWallet creates a new wallet
	CreateWallet() (Wallet, error)
	Init() error
	// CreateWalletWithIndex creates a wallet at a specific index
	CreateWalletWithIndex(addressIndex uint32) (Wallet, error)

	// GetLastAddressIndex returns the last used address index
	GetLastAddressIndex() uint32

	// SetLastAddressIndex sets the last used address index
	SetLastAddressIndex(index uint32)
}

// BaseWalletManager implements common wallet manager functionality
type BaseWalletManager struct {
	keyGenerator KeyGenerator
	coinType     uint32
}

// NewBaseWalletManager creates a new base wallet manager
func NewBaseWalletManager(generator KeyGenerator, coinType uint32) *BaseWalletManager {
	return &BaseWalletManager{
		keyGenerator: generator,
		coinType:     coinType,
	}
}

// CreatePrivateKey creates a private key at the specified index
func (m *BaseWalletManager) CreatePrivateKey(addressIndex uint32) (*ecdsa.PrivateKey, error) {
	return m.keyGenerator.CreatePrivateKey(m.coinType, 0, 0, addressIndex)
}

func (m *BaseWalletManager) GetDerivationPath(account uint32, chain uint32, addressIndex uint32) string {
	return m.keyGenerator.GetDerivationPath(m.coinType, account, chain, addressIndex)
}
