package bitcoin

import (
	"context"
	"errors"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"fiatless/internal/services/bitcoin"
	"fiatless/pkg/bitcoin/address"
	"fiatless/pkg/bitcoin/wallet"
	"fiatless/pkg/utils"
	wallet_types "fiatless/pkg/wallet"
	"fiatless/pkg/walletmgr"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletManager struct {
	wallet_types.BaseWalletManager
	blockchainService *bitcoin.Service
	repositories      *repositories.Repositories
	network           models.BitcoinNetwork
	addressType       address.BitcoinAddressType
	walletsMetaId     *uuid.UUID
	blockchainId      *uuid.UUID
	initialized       bool
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

// NewWalletManager creates a new Ethereum wallet manager
func NewWalletManager(keyGenerator wallet_types.KeyGenerator, blockchainService *bitcoin.Service, repositories *repositories.Repositories, network models.BitcoinNetwork, addressType address.BitcoinAddressType) *WalletManager {
	return &WalletManager{
		BaseWalletManager: *wallet_types.NewBaseWalletManager(keyGenerator, utils.BitcoinCoinType),
		blockchainService: blockchainService,
		repositories:      repositories,
		network:           network,
		addressType:       addressType,
	}
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

func (m *WalletManager) CreateWallet() (*wallet.BitcoinWallet, error) {
	lastAddressIndex, err := m.GetLastAddressIndex()
	if err != nil {
		return nil, err
	}
	log.Println("lastAddressIndex", lastAddressIndex)
	newAddressIndex := lastAddressIndex + 1

	wallet, err := m.createWalletWithIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}
	log.Println("newAddressIndex", newAddressIndex)
	err = m.SetLastAddressIndex(newAddressIndex)
	if err != nil {
		return nil, err
	}
	log.Println("walletsMetaId", newAddressIndex)
	walletsMetaId, err := m.getWalletsMetaId()
	if err != nil {
		return nil, err
	}
	log.Println("walletsMetaId", walletsMetaId)
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
	log.Println("walletsMetaId", walletsMetaId)
	return wallet, nil
}

func (m *WalletManager) createWalletWithIndex(addressIndex uint32) (*wallet.BitcoinWallet, error) {
	ecdsaKey, err := m.BaseWalletManager.CreatePrivateKey(addressIndex)
	if err != nil {
		return nil, err
	}

	wallet := wallet.NewBitcoinWallet(ecdsaKey, m.blockchainService, m.network, m.addressType)

	return wallet, nil
}

func (m *WalletManager) GetWallet(address address.BitcoinAddress) (*wallet.BitcoinWallet, error) {
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

func (m *WalletManager) GetMainWallet() (*wallet.BitcoinWallet, error) {
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
	log.Println("mainWallet", mainWallet.GetAddress().String())
	return walletmgr.GetLastAddressIndex(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String())
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
	return walletmgr.SetLastAddressIndex(context.Background(), m.repositories, blockchainId, mainWallet.GetAddress().String(), index)
}

func (m *WalletManager) getBlockchain() (*models.Blockchain, error) {
	return m.repositories.Blockchain.GetBlockchainByID(context.Background(), constants.GetBitcoinBlockchainID(m.network))
}

func (m *WalletManager) GetAllAddresses() ([]address.BitcoinAddress, error) {
	if m.walletsMetaId == nil {
		return nil, errors.New("walletsMetaId is nil")
	}
	wallets, err := m.repositories.Wallet.GetWalletsByMetadataID(context.Background(), *m.walletsMetaId)
	if err != nil {
		return nil, err
	}
	addresses := make([]address.BitcoinAddress, len(wallets))
	for i, wallet := range wallets {
		address, err := address.NewBitcoinAddressFromString(wallet.Address, m.network)
		if err != nil {
			return nil, err
		}
		addresses[i] = *address
	}
	return addresses, nil
}
func (m *WalletManager) IsInitialized() bool {
	return m.initialized
}

func (m *WalletManager) Withdraw(to *address.BitcoinAddress, amount decimal.Decimal) (bp_models.BitcoinWithdrawResponse, error) {
	addresses, err := m.GetAllAddresses()
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}
	var addressesStrings []string
	for _, address := range addresses {
		addressesStrings = append(addressesStrings, address.String())
	}
	utxos, err := m.repositories.UTXO.GetUTXOsByWalletAddresses(context.Background(), addressesStrings)
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}
	var utxosBp []bp_models.UTXO
	for _, utxo := range utxos {
		utxoAddress, err := address.NewBitcoinAddressFromString(utxo.Address, m.network)
		if err != nil {
			return bp_models.BitcoinWithdrawResponse{}, err
		}
		wallet, err := m.GetWallet(*utxoAddress)
		if err != nil {
			return bp_models.BitcoinWithdrawResponse{}, err
		}
		privBytes := wallet.GetPrivateKey().Key.Bytes()
		privBase58 := base58.Encode(privBytes[:])
		utxosBp = append(utxosBp, bp_models.UTXO{
			TransactionID:     utxo.Transaction.TxID,
			Vout:              utxo.Vout,
			PrivateKey:        privBase58,
			Amount:            utxo.Amount,
			ScriptPubKeyBytes: utxo.ScriptPubKeyBytes,
		})
	}
	return m.blockchainService.Withdraw(to, amount, utxosBp)
}
