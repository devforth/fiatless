package ethereum

import (
	"context"
	"errors"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"fiatless/internal/services/ethereum"
	"fiatless/pkg/ethereum/address"
	"fiatless/pkg/ethereum/wallet"
	"fiatless/pkg/utils"
	wallet_types "fiatless/pkg/wallet"
	"fiatless/pkg/walletmgr"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletManager struct {
	wallet_types.BaseWalletManager
	blockchainService *ethereum.Service
	repositories      *repositories.Repositories
	network           models.EthereumNetwork
	walletsMetaId     *uuid.UUID
	blockchainId      *uuid.UUID
	initialized       bool
}

func NewWalletManager(keyGenerator wallet_types.KeyGenerator, blockchainService *ethereum.Service, repositories *repositories.Repositories, network models.EthereumNetwork) *WalletManager {
	return &WalletManager{
		BaseWalletManager: *wallet_types.NewBaseWalletManager(keyGenerator, utils.EthereumCoinType),
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
	walletsMetaId, err := walletmgr.EnsureInit(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String(), m.GetDerivationPath(0, 0, 0))
	if err != nil {
		return err
	}
	m.walletsMetaId = &walletsMetaId
	m.initialized = true
	return nil
}

func (m *WalletManager) getWalletsMetaId() (uuid.UUID, error) {
	if m.walletsMetaId == nil {
		blockchain, err := m.getBlockchain()
		if err != nil {
			return uuid.Nil, err
		}
		mainWallet, err := m.GetMainWallet()
		if err != nil {
			return uuid.Nil, err
		}
		walletMeta, err := m.repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(context.Background(), blockchain.ID, mainWallet.GetAddress().String())
		if err != nil {
			return uuid.Nil, err
		}
		m.walletsMetaId = &walletMeta.ID
	}
	return *m.walletsMetaId, nil
}

func (m *WalletManager) getBlockchainId() (uuid.UUID, error) {
	if m.blockchainId == nil {
		blockchain, err := m.getBlockchain()
		if err != nil {
			return uuid.Nil, err
		}
		m.blockchainId = &blockchain.ID
	}
	return *m.blockchainId, nil
}

func (m *WalletManager) CreateWallet() (*wallet.EthereumWallet, error) {
	lastAddressIndex, err := m.GetLastAddressIndex()
	if err != nil {
		return nil, err
	}

	newAddressIndex := lastAddressIndex + 1

	wallet, err := m.createWalletWithIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}

	err = m.SetLastAddressIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}

	walletsMetaId, err := m.getWalletsMetaId()
	if err != nil {
		return nil, err
	}

	err = m.repositories.Wallet.CreateWallet(context.Background(), &models.Wallet{
		ID:             uuid.New(),
		Address:        wallet.GetAddress().String(),
		Index:          newAddressIndex,
		DerivationPath: m.GetDerivationPath(0, 0, newAddressIndex),
		MetaID:         walletsMetaId,
	})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m *WalletManager) createWalletWithIndex(addressIndex uint32) (*wallet.EthereumWallet, error) {
	ecdsaKey, err := m.BaseWalletManager.CreatePrivateKey(addressIndex)
	if err != nil {
		return nil, err
	}

	wallet := wallet.NewEthereumWallet(ecdsaKey, m.blockchainService)

	return wallet, nil
}

func (m *WalletManager) GetWallet(address address.EthereumAddress) (*wallet.EthereumWallet, error) {
	dbWallet, err := m.repositories.Wallet.GetWalletByAddress(context.Background(), address.String())
	if err != nil {
		wallet, err := m.GetMainWallet()
		if err != nil {
			return nil, err
		}
		if address.String() == wallet.GetAddress().String() {
			return wallet, nil
		}
		return nil, errors.New("wallet not found")
	}
	return m.createWalletWithIndex(dbWallet.Index)
}

func (m *WalletManager) GetMainWallet() (*wallet.EthereumWallet, error) {
	return m.createWalletWithIndex(0)
}

func (m *WalletManager) GetLastAddressIndex() (uint32, error) {
	blockchainId, err := m.getBlockchainId()
	if err != nil {
		return 0, err
	}
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return 0, err
	}
	walletMeta, err := m.repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(context.Background(), blockchainId, mainWallet.GetAddress().String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = m.repositories.WalletMeta.CreateWalletMeta(context.Background(), &models.WalletMeta{
				ID:           uuid.New(),
				MainWallet:   mainWallet.GetAddress().String(),
				BlockchainID: blockchainId,
			})
			if err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, err
	}
	return walletMeta.LastIndex, nil
}

func (m *WalletManager) SetLastAddressIndex(index uint32) error {
	blockchainId, err := m.getBlockchainId()
	if err != nil {
		return err
	}
	mainWallet, err := m.GetMainWallet()
	if err != nil {
		return err
	}
	walletMeta, err := m.repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(context.Background(), blockchainId, mainWallet.GetAddress().String())
	if err != nil {
		return err
	}
	walletMeta.LastIndex = index
	return m.repositories.WalletMeta.UpdateWalletMeta(context.Background(), walletMeta)
}

func (m *WalletManager) getBlockchain() (*models.Blockchain, error) {
	return m.repositories.Blockchain.GetBlockchainBySymbolAndNetwork(context.Background(), "ETH", models.BlockchainNetwork(m.network))
}
