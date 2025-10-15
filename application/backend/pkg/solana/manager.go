package solana

import (
	"context"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	solana_svc "fiatless/internal/services/solana"
	"fiatless/pkg/solana/address"
	solana_wallet "fiatless/pkg/solana/wallet"
	"fiatless/pkg/utils"
	"fiatless/pkg/wallet"
	"fiatless/pkg/walletmgr"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type WalletManager struct {
	repositories  *repositories.Repositories
	network       models.SolanaNetwork
	walletsMetaId *uuid.UUID
	blockchainId  *uuid.UUID
	initialized   bool
	service       *solana_svc.Service
	keyGenerator  wallet.Ed25519KeyGenerator
}

func NewWalletManager(generator wallet.Ed25519KeyGenerator, service *solana_svc.Service, repositories *repositories.Repositories, network models.SolanaNetwork) *WalletManager {
	return &WalletManager{repositories: repositories, network: network, service: service, keyGenerator: generator}
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
	metaId, err := walletmgr.EnsureInit(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String(), "m/44'/501'/0'/0/0")
	if err != nil {
		return err
	}
	m.walletsMetaId = &metaId
	m.initialized = true
	return nil
}

func (m *WalletManager) CreateWallet() (*solana_wallet.SolanaWallet, error) {
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
		DerivationPath: "m/44'/501'/0'/0/" + fmt.Sprint(newAddressIndex),
		MetaID:         *m.walletsMetaId,
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
func (m *WalletManager) createWalletWithIndex(index uint32) (*solana_wallet.SolanaWallet, error) {
	priv, err := m.keyGenerator.CreateEd25519PrivateKey(utils.SolanaCoinType, 0, 0, index)
	if err != nil {
		return nil, err
	}
	return solana_wallet.NewSolanaWallet(&priv, m.service), nil
}

func (m *WalletManager) getBlockchainId() (uuid.UUID, error) {
	if m.blockchainId == nil {
		blockchainId := constants.GetSolanaBlockchainID(m.network)
		m.blockchainId = &blockchainId
	}
	return *m.blockchainId, nil
}

func (m *WalletManager) GetWallet(addr address.SolanaAddress) (*solana_wallet.SolanaWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	dbWallet, err := m.repositories.Wallet.GetWalletByAddress(context.Background(), addr.String())
	if err != nil {
		wallet, err := m.GetMainWallet()
		if err != nil {
			return nil, err
		}
		if wallet.GetAddress().String() == addr.String() {
			return wallet, nil
		}
		return nil, errors.New("wallet not found")
	}
	return m.createWalletWithIndex(dbWallet.Index)
}

func (m *WalletManager) GetWalletByIndex(index uint32) (*solana_wallet.SolanaWallet, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	return m.createWalletWithIndex(index)
}

func (m *WalletManager) GetMainWallet() (*solana_wallet.SolanaWallet, error) {
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

func (m *WalletManager) GetAllAddresses() ([]address.SolanaAddress, error) {
	if !m.initialized {
		return nil, errors.New("wallet manager not initialized")
	}
	log.Println("Getting wallets by metadata id", *m.walletsMetaId)
	wallets, err := m.repositories.Wallet.GetWalletsByMetadataID(context.Background(), *m.walletsMetaId)
	if err != nil {
		return nil, err
	}
	addresses := make([]address.SolanaAddress, len(wallets))
	for i, wallet := range wallets {
		address, err := address.NewSolanaAddressFromString(wallet.Address)
		if err != nil {
			return nil, err
		}
		addresses[i] = *address
	}
	return addresses, nil
}

func (m *WalletManager) GetWallets(excludeMainWallet bool) ([]*solana_wallet.SolanaWallet, error) {
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

	var wallets []*solana_wallet.SolanaWallet
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
