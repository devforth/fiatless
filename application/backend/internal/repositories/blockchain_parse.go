package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockchainParseRepository interface {
	GetBlockchainParse(ctx context.Context, blockchainID uuid.UUID) (*models.BlockchainParse, error)
	UpdateBlockchainParse(ctx context.Context, blockchainID uuid.UUID, lastBlockNumber uint64) error
	CreateBlockchainParse(ctx context.Context, blockchainID uuid.UUID, lastBlockNumber *uint64) error
}

type SQLBlockchainParseRepository struct {
	db *gorm.DB
}

func NewSQLBlockchainParseRepository(db *gorm.DB) BlockchainParseRepository {
	return &SQLBlockchainParseRepository{db: db}
}

func (r *SQLBlockchainParseRepository) GetBlockchainParse(ctx context.Context, blockchainID uuid.UUID) (*models.BlockchainParse, error) {
	var blockchainParse models.BlockchainParse
	result := r.db.WithContext(ctx).Model(&models.BlockchainParse{}).Where("blockchain_id = ?", blockchainID).First(&blockchainParse)
	if result.Error != nil {
		return nil, result.Error
	}
	return &blockchainParse, nil
}

func (r *SQLBlockchainParseRepository) UpdateBlockchainParse(ctx context.Context, blockchainID uuid.UUID, lastBlockNumber uint64) error {
	result := r.db.WithContext(ctx).Model(&models.BlockchainParse{}).Where("blockchain_id = ?", blockchainID).Update("last_block_number", lastBlockNumber)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *SQLBlockchainParseRepository) CreateBlockchainParse(ctx context.Context, blockchainID uuid.UUID, lastBlockNumber *uint64) error {
	blockchainParse := models.BlockchainParse{
		BlockchainID:    blockchainID,
		LastBlockNumber: lastBlockNumber,
	}
	return r.db.WithContext(ctx).Create(&blockchainParse).Error
}
