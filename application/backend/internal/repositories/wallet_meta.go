package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletMetaRepository interface {
	GetWalletMetaByBlockchainIDAndWalletAddress(ctx context.Context, blockchainID uuid.UUID, walletAddress string) (*models.WalletMeta, error)
	CreateWalletMeta(ctx context.Context, walletMeta *models.WalletMeta) error
	UpdateWalletMeta(ctx context.Context, walletMeta *models.WalletMeta) error
}

type SQLWalletMetaRepository struct {
	db *gorm.DB
}

func NewSQLWalletMetaRepository(db *gorm.DB) WalletMetaRepository {
	return &SQLWalletMetaRepository{
		db: db,
	}
}

func (r *SQLWalletMetaRepository) GetWalletMetaByBlockchainIDAndWalletAddress(ctx context.Context, blockchainID uuid.UUID, walletAddress string) (*models.WalletMeta, error) {
	var walletMeta models.WalletMeta
	result := r.db.WithContext(ctx).Where("blockchain_id = ? AND main_wallet = ?", blockchainID, walletAddress).First(&walletMeta)
	return &walletMeta, result.Error
}

func (r *SQLWalletMetaRepository) CreateWalletMeta(ctx context.Context, walletMeta *models.WalletMeta) error {
	return r.db.WithContext(ctx).Create(walletMeta).Error
}

func (r *SQLWalletMetaRepository) UpdateWalletMeta(ctx context.Context, walletMeta *models.WalletMeta) error {
	return r.db.WithContext(ctx).Save(walletMeta).Error
}
