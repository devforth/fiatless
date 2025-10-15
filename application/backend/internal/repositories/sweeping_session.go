package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SweepingSessionRepository interface {
	CreateSweepingSession(ctx context.Context, sweepingSession *models.SweepingSession) (*models.SweepingSession, error)
	UpdateSweepingSessionStatus(ctx context.Context, sweepingSessionID uuid.UUID, status models.SweepingSessionStatus) error
}

type SQLSweepingSessionRepository struct {
	db *gorm.DB
}

func NewSQLSweepingSessionRepository(db *gorm.DB) SweepingSessionRepository {
	return &SQLSweepingSessionRepository{
		db: db,
	}
}

func (r *SQLSweepingSessionRepository) CreateSweepingSession(ctx context.Context, sweepingSession *models.SweepingSession) (*models.SweepingSession, error) {
	result := r.db.WithContext(ctx).Create(&sweepingSession)
	if result.Error != nil {
		return nil, result.Error
	}
	return sweepingSession, nil
}

func (r *SQLSweepingSessionRepository) UpdateSweepingSessionStatus(ctx context.Context, sweepingSessionID uuid.UUID, status models.SweepingSessionStatus) error {
	result := r.db.WithContext(ctx).Model(&models.SweepingSession{}).Where("id = ?", sweepingSessionID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
