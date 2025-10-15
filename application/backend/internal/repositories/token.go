package repositories

import (
	"context"
	"fiatless/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository interface {
	GetTokensByType(ctx context.Context, tokenType models.TokenType) ([]models.Token, error)
	GetTokensByTypeAndBlockchain(ctx context.Context, tokenType models.TokenType, blockchainID uuid.UUID) ([]models.Token, error)
	GetToken(ctx context.Context, token *models.Token) (*models.Token, error)
	GetTokens(ctx context.Context, token *models.Token) ([]models.Token, error)
	GetTokensByBlockchains(ctx context.Context, token *models.Token, blockchainIDs []uuid.UUID) ([]models.Token, error)
	GetAllTokens(ctx context.Context, token *models.Token) ([]models.Token, error)
	GetTokenByTokenID(ctx context.Context, tokenID *string) (*models.Token, error)
	AddToken(ctx context.Context, token *models.Token) error
}

type SQLTokenRepository struct {
	db *gorm.DB
}

func NewSQLTokenRepository(db *gorm.DB) TokenRepository {
	return &SQLTokenRepository{
		db: db,
	}
}

func (r *SQLTokenRepository) GetToken(ctx context.Context, token *models.Token) (*models.Token, error) {
	var tokenResult *models.Token

	result := r.db.WithContext(ctx).Where(token).First(&tokenResult)
	if result.Error != nil {
		return nil, result.Error
	}

	return tokenResult, nil
}

func (r *SQLTokenRepository) GetTokens(ctx context.Context, token *models.Token) ([]models.Token, error) {
	var tokens []models.Token

	result := r.db.WithContext(ctx).Where(token).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}

	return tokens, nil
}

func (r *SQLTokenRepository) GetTokensByBlockchains(ctx context.Context, token *models.Token, blockchainIDs []uuid.UUID) ([]models.Token, error) {
	var tokens []models.Token

	query := r.db.WithContext(ctx).Where(token).Where("blockchain_id IN ?", blockchainIDs)

	result := query.Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}

	return tokens, nil
}

func (r *SQLTokenRepository) GetAllTokens(ctx context.Context, token *models.Token) ([]models.Token, error) {
	var tokens []models.Token

	result := r.db.WithContext(ctx).Where(token).Find(&tokens)

	if result.Error != nil {
		return nil, result.Error
	}

	return tokens, nil
}

func (r *SQLTokenRepository) GetTokensByType(ctx context.Context, tokenType models.TokenType) ([]models.Token, error) {
	return r.GetTokens(ctx, &models.Token{Type: tokenType})
}

func (r *SQLTokenRepository) GetTokensByTypeAndBlockchain(ctx context.Context, tokenType models.TokenType, blockchainID uuid.UUID) ([]models.Token, error) {
	return r.GetTokens(ctx, &models.Token{Type: tokenType, BlockchainID: blockchainID})
}

func (r *SQLTokenRepository) GetTokenByTokenID(ctx context.Context, tokenID *string) (*models.Token, error) {
	return r.GetToken(ctx, &models.Token{TokenID: tokenID})
}

func (r *SQLTokenRepository) AddToken(ctx context.Context, token *models.Token) error {
	return r.db.WithContext(ctx).Create(token).Error
}
