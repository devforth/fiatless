package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetBalance(ctx context.Context, address string, token_id string) (*decimal.Decimal, error)
	GetLastTransactionId(ctx context.Context, address string) (*string, error)
}

type SQLTransactionRepository struct {
	db *gorm.DB
}

func NewSQLTransactionRepository(db *gorm.DB) TransactionRepository {
	return &SQLTransactionRepository{
		db: db,
	}
}

func (r *SQLTransactionRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *SQLTransactionRepository) GetBalance(ctx context.Context, address string, token_id string) (*decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("to_address = ? AND token_id = ?", address, token_id).Select("COALESCE(SUM(amount), 0)").Scan(&balance).Error
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *SQLTransactionRepository) GetLastTransactionId(ctx context.Context, address string) (*string, error) {
	var transaction models.Transaction
	err := r.db.WithContext(ctx).Where("to_address = ?", address).Order("created_at DESC").First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction.TxID, nil
}
