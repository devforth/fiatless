package tron

import (
	"context"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"fiatless/internal/services/tron"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/wallet"
	"fiatless/pkg/utils"
	wallet_types "fiatless/pkg/wallet"
	"fiatless/pkg/walletmgr"
	"log"

	"github.com/google/uuid"
)

type WalletManager struct {
	wallet_types.BaseWalletManager
	blockchainService *tron.Service
	repositories      *repositories.Repositories
	network           models.TronNetwork
	walletsMetaId     *uuid.UUID
	blockchainId      *uuid.UUID
	initialized       bool
}

func NewWalletManager(keyGenerator wallet_types.KeyGenerator, blockchainService *tron.Service, repositories *repositories.Repositories, network models.TronNetwork) *WalletManager {
	return &WalletManager{
		BaseWalletManager: *wallet_types.NewBaseWalletManager(keyGenerator, utils.TronCoinType),
		blockchainService: blockchainService,
		repositories:      repositories,
		network:           network,
	}
}

func (m *WalletManager) Init() error {
	if m.initialized {
		return nil
	}
	blockchainId, err := m.getBlockchainId()
	if err != nil {
		return err
	}
	m.blockchainId = &blockchainId
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return err
	}
	metaId, err := walletmgr.EnsureInit(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String(), m.GetDerivationPath(0, 0, 0))
	if err != nil {
		return err
	}
	m.walletsMetaId = &metaId
	m.initialized = true
	return nil
}

func (m *WalletManager) CreateWallet() (*wallet.TronWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	lastAddressIndex, err := m.getLastAddressIndex()
	if err != nil {
		return nil, err
	}

	newAddressIndex := lastAddressIndex + 1

	wallet, err := m.createWalletWithIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}

	err = m.setLastAddressIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}

	err = m.repositories.Wallet.CreateWallet(context.Background(), &models.Wallet{
		ID:             uuid.New(),
		Address:        wallet.GetAddress().String(),
		Index:          newAddressIndex,
		DerivationPath: m.GetDerivationPath(0, 0, newAddressIndex),
		MetaID:         *m.walletsMetaId,
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
func (m *WalletManager) createWalletWithIndex(addressIndex uint32) (*wallet.TronWallet, error) {
	ecdsaKey, err := m.BaseWalletManager.CreatePrivateKey(addressIndex)
	if err != nil {
		return nil, err
	}

	wallet := wallet.NewTronWallet(ecdsaKey, m.blockchainService)
	return wallet, nil
}

func (m *WalletManager) getBlockchainId() (uuid.UUID, error) {
	if m.blockchainId == nil {
		blockchainId := constants.GetTronBlockchainID(m.network)
		m.blockchainId = &blockchainId
	}
	return *m.blockchainId, nil
}

func (m *WalletManager) GetWallet(address address.TronAddress) (*wallet.TronWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	dbWallet, err := m.repositories.Wallet.GetWalletByAddress(context.Background(), address.String())
	if err != nil {
		wallet, err := m.GetMainWallet()
		if err != nil {
			return nil, err
		}
		if wallet.GetAddress().String() == address.String() {
			return wallet, nil
		}
		return nil, errors.New("wallet not found")
	}
	return m.createWalletWithIndex(dbWallet.Index)
}

func (m *WalletManager) GetWalletByIndex(index uint32) (*wallet.TronWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	ecdsaKey, err := m.BaseWalletManager.CreatePrivateKey(index)
	if err != nil {
		return nil, err
	}
	return wallet.NewTronWallet(ecdsaKey, m.blockchainService), nil
}

func (m *WalletManager) GetMainWallet() (*wallet.TronWallet, error) {
	return m.createWalletWithIndex(0)
}

func (m *WalletManager) getLastAddressIndex() (uint32, error) {
	blockchainId, err := m.getBlockchainId()
	if err != nil {
		return 0, err
	}
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return 0, err
	}
	return walletmgr.GetLastAddressIndex(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String())
}

func (m *WalletManager) setLastAddressIndex(index uint32) error {
	blockchainId, err := m.getBlockchainId()
	if err != nil {
		return err
	}
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return err
	}
	return walletmgr.SetLastAddressIndex(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String(), index)
}

func (m *WalletManager) GetAllAddresses() ([]address.TronAddress, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	log.Println("Getting wallets by metadata id", *m.walletsMetaId)
	wallets, err := m.repositories.Wallet.GetWalletsByMetadataID(context.Background(), *m.walletsMetaId)
	if err != nil {
		return nil, err
	}
	addresses := make([]address.TronAddress, len(wallets))
	for i, wallet := range wallets {
		address, err := address.NewTronAddressFromBase58(wallet.Address)
		if err != nil {
			return nil, err
		}
		addresses[i] = *address
	}
	return addresses, nil
}

func (m *WalletManager) GetWallets(excludeMainWallet bool) ([]*wallet.TronWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return nil, err
	}
	addresses, err := m.GetAllAddresses()
	if err != nil {
		return nil, err
	}

	var wallets []*wallet.TronWallet
	for _, address := range addresses {
		if excludeMainWallet && address.String() == mainWallet.GetAddress().String() {
			continue
		}
		wallet, err := m.GetWallet(address)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}

func (m *WalletManager) IsInitialized() bool {
	return m.initialized
}
