package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UTXORepository interface {
	GetUTXOsByWalletAddresses(ctx context.Context, walletAddresses []string) ([]models.UTXO, error)
	DeleteUTXOByTxIDWithVout(ctx context.Context, txID string, vout int) error
	DeleteUTXO(ctx context.Context, utxo *models.UTXO) error
	CreateUTXO(ctx context.Context, utxo *models.UTXO) error
	GetBalance(ctx context.Context, address string) (*decimal.Decimal, error)
}

type SQLUTXORepository struct {
	db *gorm.DB
}

func NewSQLUTXORepository(db *gorm.DB) UTXORepository {
	return &SQLUTXORepository{db: db}
}

func (r *SQLUTXORepository) GetUTXOsByWalletAddresses(ctx context.Context, walletAddresses []string) ([]models.UTXO, error) {
	var utxos []models.UTXO
	result := r.db.WithContext(ctx).Model(&models.UTXO{}).Preload("Transaction").Where("address IN ?", walletAddresses).Find(&utxos)
	if result.Error != nil {
		return nil, result.Error
	}
	return utxos, nil
}

func (r *SQLUTXORepository) DeleteUTXOByTxIDWithVout(ctx context.Context, txID string, vout int) error {
	var transaction models.Transaction
	err := r.db.WithContext(ctx).Where("tx_id = ?", txID).First(&transaction).Error
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&models.UTXO{}).Where("transaction_id = ? AND vout = ?", transaction.ID, vout).Delete(&models.UTXO{}).Error
}

func (r *SQLUTXORepository) DeleteUTXO(ctx context.Context, utxo *models.UTXO) error {
	return r.db.WithContext(ctx).Delete(utxo).Error
}

func (r *SQLUTXORepository) CreateUTXO(ctx context.Context, utxo *models.UTXO) error {
	return r.db.WithContext(ctx).Create(utxo).Error
}

func (r *SQLUTXORepository) GetBalance(ctx context.Context, address string) (*decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.db.WithContext(ctx).Model(&models.UTXO{}).Where("address = ?", address).Select("COALESCE(SUM(amount), 0)").Scan(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}
