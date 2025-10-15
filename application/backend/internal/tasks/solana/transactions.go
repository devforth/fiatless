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
	"gorm.io/gorm"
)

type TransactionsTask struct {
	BlockchainID  uuid.UUID
	service       *solana_service.Service
	walletManager *solana.WalletManager
	repositories  *repositories.Repositories
	network       models.SolanaNetwork
}

func NewTransactionsTask(blockchainID uuid.UUID, service *solana_service.Service, walletManager *solana.WalletManager, repositories *repositories.Repositories) *TransactionsTask {
	return &TransactionsTask{
		BlockchainID:  blockchainID,
		service:       service,
		walletManager: walletManager,
		repositories:  repositories,
		network:       constants.GetSolanaNetByBlockchainID(blockchainID),
	}
}

func (t *TransactionsTask) Do(ctx context.Context) error {
	log.Println("SolanaTransactions: fetching addresses")
	addresses, err := t.walletManager.GetAllAddresses()
	if err != nil {
		return err
	}

	for len(addresses) > 0 {
		next := addresses[:0]
		for _, address := range addresses {
			lastTransactionId, err := t.repositories.Transaction.GetLastTransactionId(ctx, address.String())
			log.Println("Last transaction id", lastTransactionId)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					lastTransactionId = nil
				} else {
					return err
				}
			}
			transactions, err := t.service.GetWalletTransactions(address, lastTransactionId)
			if err != nil {
				return err
			}
			for _, transaction := range transactions.Transactions {
				transactionType := models.TransactionTypeDeposit
				if transaction.Amount.IsNegative() {
					transactionType = models.TransactionTypeWithdraw
				}
				t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
					ID:        uuid.New(),
					TxID:      transaction.TxID,
					TokenID:   uuid.MustParse(constants.GetSolanaTokenID(t.network)),
					ToAddress: transaction.Address,
					Amount:    transaction.Amount,
					Fee:       transaction.Fee,
					Type:      transactionType,
					CreatedAt: time.Unix(transaction.Timestamp, 0),
				})
			}
			if len(transactions.Transactions) == constants.SolanaLimitTransactions {
				next = append(next, address)
			}
		}
		addresses = next
	}
	return nil
}
