package solana

import (
	"context"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	solana_service "fiatless/internal/services/solana"
	"fiatless/pkg/solana"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BlocksTask struct {
	BlockchainID  uuid.UUID
	service       *solana_service.Service
	walletManager *solana.WalletManager
	repositories  *repositories.Repositories
}

func NewBlocksTask(blockchainID uuid.UUID, service *solana_service.Service, walletManager *solana.WalletManager, repositories *repositories.Repositories) *BlocksTask {
	return &BlocksTask{
		BlockchainID:  blockchainID,
		service:       service,
		walletManager: walletManager,
		repositories:  repositories,
	}
}

func (t *BlocksTask) Do(ctx context.Context) error {
	log.Println("SolanaBlocks: fetching addresses")
	wallets, err := t.walletManager.GetAllAddresses()
	if err != nil {
		return err
	}

	// Read last parsed block
	blockchainParse, err := t.repositories.BlockchainParse.GetBlockchainParse(ctx, t.BlockchainID)
	var lastBlockNumber *uint64
	if err == nil {
		lastBlockNumber = blockchainParse.LastBlockNumber
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := t.repositories.BlockchainParse.CreateBlockchainParse(ctx, t.BlockchainID, nil); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	network := constants.GetSolanaNetByBlockchainID(t.BlockchainID)
	// Resolve native SOL token ID from DB (BlockchainID + Native type)
	nativeToken := constants.GetSolanaTokenID(network)

	// Call service to parse blocks
	resp, err := t.service.ParseBlocks(wallets, nativeToken, lastBlockNumber)
	if err != nil {
		return err
	}

	// Persist deposits as transactions
	for _, dep := range resp.Deposits {
		amt, _ := decimal.NewFromString(dep.Amount)
		t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
			ID:        uuid.New(),
			TxID:      dep.TxID,
			TokenID:   uuid.MustParse(nativeToken),
			ToAddress: dep.ToAddress,
			Amount:    amt,
			Fee:       decimal.Zero,
			Type:      models.TransactionTypeDeposit,
			CreatedAt: time.Unix(dep.Timestamp, 0),
		})
	}

	// Update last parsed block number
	if err := t.repositories.BlockchainParse.UpdateBlockchainParse(ctx, t.BlockchainID, resp.LastBlockNumber); err != nil {
		return err
	}
	return nil
}
