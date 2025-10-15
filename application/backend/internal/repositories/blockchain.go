package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockchainRepository interface {
	GetBlockchain(ctx context.Context, blockchain *models.Blockchain) (*models.Blockchain, error)
	GetBlockchainByID(ctx context.Context, id uuid.UUID) (*models.Blockchain, error)
	GetBlockchainBySymbol(ctx context.Context, symbol string) (*models.Blockchain, error)
	GetBlockchainBySymbolAndNetwork(ctx context.Context, symbol string, network models.BlockchainNetwork) (*models.Blockchain, error)
	AddBlockchain(ctx context.Context, blockchain *models.Blockchain) error
}

type SQLBlockchainRepository struct {
	db *gorm.DB
}

func NewSQLBlockchainRepository(db *gorm.DB) BlockchainRepository {
	return &SQLBlockchainRepository{
		db: db,
	}
}

func (r *SQLBlockchainRepository) GetBlockchain(ctx context.Context, blockchain *models.Blockchain) (*models.Blockchain, error) {
	var blockchainResult models.Blockchain

	result := r.db.WithContext(ctx).Model(&models.Blockchain{}).Where(blockchain).First(&blockchainResult)
	if result.Error != nil {
		return nil, result.Error
	}

	return &blockchainResult, nil
}

func (r *SQLBlockchainRepository) GetBlockchainByID(ctx context.Context, id uuid.UUID) (*models.Blockchain, error) {
	return r.GetBlockchain(ctx, &models.Blockchain{ID: id})
}

func (r *SQLBlockchainRepository) GetBlockchainBySymbol(ctx context.Context, symbol string) (*models.Blockchain, error) {
	return r.GetBlockchain(ctx, &models.Blockchain{Symbol: symbol})
}

func (r *SQLBlockchainRepository) GetBlockchainBySymbolAndNetwork(ctx context.Context, symbol string, network models.BlockchainNetwork) (*models.Blockchain, error) {
	return r.GetBlockchain(ctx, &models.Blockchain{Symbol: symbol, Network: network})
}

func (r *SQLBlockchainRepository) AddBlockchain(ctx context.Context, blockchain *models.Blockchain) error {
	return r.db.WithContext(ctx).Create(blockchain).Error
}
