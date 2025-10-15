package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletRepository interface {
	GetWalletByAddress(ctx context.Context, address string) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	GetWalletsByMetadataID(ctx context.Context, metadataID uuid.UUID) ([]*models.Wallet, error)
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
	UpdateWallet(ctx context.Context, wallet *models.Wallet) error
	DeleteWallet(ctx context.Context, wallet *models.Wallet) error
	Balance(ctx context.Context, address string, tokenID uuid.UUID) (*decimal.Decimal, error)
}

type SQLWalletRepository struct {
	db *gorm.DB
}

func NewSQLWalletRepository(db *gorm.DB) WalletRepository {
	return &SQLWalletRepository{
		db: db,
	}
}

func (r *SQLWalletRepository) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	result := r.db.Where("id = ?", id).First(&wallet)
	return &wallet, result.Error
}

func (r *SQLWalletRepository) GetWalletsByMetadataID(ctx context.Context, metadataID uuid.UUID) ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	result := r.db.Where("meta_id = ?", metadataID).Find(&wallets)
	return wallets, result.Error
}

func (r *SQLWalletRepository) GetWalletByAddress(ctx context.Context, address string) (*models.Wallet, error) {
	var wallet models.Wallet
	result := r.db.Where("address = ?", address).First(&wallet)
	return &wallet, result.Error
}

func (r *SQLWalletRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *SQLWalletRepository) UpdateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *SQLWalletRepository) DeleteWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.Delete(wallet).Error
}

func (r *SQLWalletRepository) Balance(ctx context.Context, address string, tokenID uuid.UUID) (*decimal.Decimal, error) {
	var balance decimal.Decimal
	result := r.db.Raw("SELECT COALESCE(SUM(amount), 0) as amount FROM transactions WHERE to_address = ? AND token_id = ?", address, tokenID).Scan(&balance)
	return &balance, result.Error
}
