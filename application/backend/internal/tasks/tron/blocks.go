package tron

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	tron_service "fiatless/internal/services/tron"
	"fiatless/pkg/tron"
)

type BlocksTask struct {
	BlockchainID  uuid.UUID
	service       *tron_service.Service
	walletManager *tron.WalletManager
	repositories  *repositories.Repositories
}

func NewBlocksTask(blockchainID uuid.UUID, service *tron_service.Service, walletManager *tron.WalletManager, repositories *repositories.Repositories) *BlocksTask {
	return &BlocksTask{
		BlockchainID:  blockchainID,
		service:       service,
		walletManager: walletManager,
		repositories:  repositories,
	}
}

func (t *BlocksTask) Do(ctx context.Context) error {
	log.Println("Getting all addresses")
	wallets, err := t.walletManager.GetAllAddresses()
	if err != nil {
		return err
	}
	blockchainParse, err := t.repositories.BlockchainParse.GetBlockchainParse(ctx, t.BlockchainID)
	var lastBlockNumber *uint64 = nil
	if err == nil {
		lastBlockNumber = blockchainParse.LastBlockNumber
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = t.repositories.BlockchainParse.CreateBlockchainParse(ctx, t.BlockchainID, nil)
			if err != nil {
				return err
			}
		}
	}
	tokens := []bp_models.ParseBlocksToken{}
	allTokens, err := t.repositories.Token.GetTokens(ctx, &models.Token{
		BlockchainID: t.BlockchainID,
		IsActive:     true,
	})
	if err != nil {
		return err
	}
	for _, token := range allTokens {
		tokens = append(tokens, bp_models.ParseBlocksToken{
			ID:        token.ID,
			TokenID:   token.TokenID,
			TokenType: token.Type,
		})
	}
	response, err := t.service.ParseBlocks(wallets, tokens, lastBlockNumber)
	if err != nil {
		return err
	}
	log.Println("Updating blockchain parse", response.LastBlockNumber)
	for _, deposit := range response.Deposits {
		amount, _ := decimal.NewFromString(deposit.Amount)
		t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
			ID:        uuid.New(),
			TxID:      deposit.TxID,
			TokenID:   uuid.MustParse(deposit.TokenID),
			Fee:       decimal.Zero,
			Type:      models.TransactionTypeDeposit,
			ToAddress: deposit.ToAddress,
			Amount:    amount,
			CreatedAt: time.Unix(deposit.Timestamp, 0),
		})
	}
	err = t.repositories.BlockchainParse.UpdateBlockchainParse(ctx, t.BlockchainID, response.LastBlockNumber)
	if err != nil {
		return err
	}
	return nil
}
