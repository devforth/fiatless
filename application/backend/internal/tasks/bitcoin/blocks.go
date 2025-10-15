package bitcoin

import (
	"context"
	"encoding/hex"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	bitcoin_service "fiatless/internal/services/bitcoin"
	"fiatless/pkg/bitcoin"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BlocksTask struct {
	BlockchainID  uuid.UUID
	service       *bitcoin_service.Service
	walletManager *bitcoin.WalletManager
	repositories  *repositories.Repositories
}

func NewBlocksTask(blockchainID uuid.UUID, service *bitcoin_service.Service, walletManager *bitcoin.WalletManager, repositories *repositories.Repositories) *BlocksTask {
	return &BlocksTask{
		BlockchainID:  blockchainID,
		service:       service,
		walletManager: walletManager,
		repositories:  repositories,
	}
}

func (t *BlocksTask) Do(ctx context.Context) error {
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

	walletAddresses := make([]string, len(wallets))
	for i, wallet := range wallets {
		walletAddresses[i] = wallet.String()
	}
	existingUTXOs, err := t.repositories.UTXO.GetUTXOsByWalletAddresses(ctx, walletAddresses)
	if err != nil {
		return err
	}

	utxos := map[string]struct{}{}
	for _, utxo := range existingUTXOs {
		utxos[utxo.Transaction.TxID+"-"+strconv.Itoa(utxo.Vout)] = struct{}{}
	}
	response, err := t.service.ParseBlocks(wallets, utxos, lastBlockNumber)
	if err != nil {
		return err
	}
	network := constants.GetBitcoinNetByBlockchainID(t.BlockchainID)
	for _, deposit := range response.Transactions {
		transactionId := uuid.New()
		if len(deposit.Vin) > 0 && len(deposit.Vout) > 0 {
			t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
				ID:        transactionId,
				TxID:      deposit.TxID,
				TokenID:   uuid.MustParse(constants.GetBitcoinTokenID(network)),
				ToAddress: "0x",         // TODO: change to null
				Amount:    decimal.Zero, // TODO: change to null
				Fee:       decimal.NewFromFloat(deposit.Fee),
				Type:      models.TransactionTypeTransfer,
				CreatedAt: time.Unix(deposit.Time, 0),
			})
		} else if len(deposit.Vout) > 0 {
			t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
				ID:        transactionId,
				TxID:      deposit.TxID,
				TokenID:   uuid.MustParse(constants.GetBitcoinTokenID(network)),
				ToAddress: "0x",         // TODO: change to null
				Amount:    decimal.Zero, // TODO: change to null
				Fee:       decimal.NewFromFloat(deposit.Fee),
				Type:      models.TransactionTypeDeposit,
				CreatedAt: time.Unix(deposit.Time, 0),
			})
		} else {
			t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
				ID:        transactionId,
				TxID:      deposit.TxID,
				TokenID:   uuid.MustParse(constants.GetBitcoinTokenID(network)),
				ToAddress: "0x",         // TODO: change to null
				Amount:    decimal.Zero, // TODO: change to null
				Fee:       decimal.NewFromFloat(deposit.Fee),
				Type:      models.TransactionTypeWithdraw,
				CreatedAt: time.Unix(deposit.Time, 0),
			})
		}
		for _, vout := range deposit.Vout {
			amount, _ := decimal.NewFromString(vout.Amount)

			scriptPubKeyBytes, err := hex.DecodeString(vout.Scriptpubkeyhex)
			if err != nil {
				return err
			}
			t.repositories.UTXO.CreateUTXO(ctx, &models.UTXO{
				ID:                uuid.New(),
				TransactionID:     transactionId,
				Address:           vout.Address,
				Vout:              vout.Vout,
				ScriptPubKeyBytes: scriptPubKeyBytes,
				Amount:            amount,
			})
		}
		for _, vin := range deposit.Vin {
			t.repositories.UTXO.DeleteUTXOByTxIDWithVout(ctx, vin.TxID, vin.Vout)
		}
	}

	err = t.repositories.BlockchainParse.UpdateBlockchainParse(ctx, t.BlockchainID, response.LastBlockNumber)
	if err != nil {
		return err
	}
	return nil
}
