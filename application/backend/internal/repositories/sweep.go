package repositories

import (
	"context"
	"fiatless/internal/models"

	"gorm.io/gorm"
)

type SweepRepository interface {
	CreateSweep(ctx context.Context, sweep *models.Sweep) (*models.Sweep, error)
}

type SQLSweepRepository struct {
	db *gorm.DB
}

func NewSQLSweepRepository(db *gorm.DB) SweepRepository {
	return &SQLSweepRepository{
		db: db,
	}
}

func (r *SQLSweepRepository) CreateSweep(ctx context.Context, sweep *models.Sweep) (*models.Sweep, error) {
	result := r.db.WithContext(ctx).Create(&sweep)
	if result.Error != nil {
		return nil, result.Error
	}
	return sweep, nil
}
