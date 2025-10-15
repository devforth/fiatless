package routes

import (
	"fiatless/internal/models"
	"time"

	"github.com/google/uuid"
)

type GetTokensRequest struct {
	IncludeInactive bool `in:"query=includeInactive"`
}

type GetTokenRequest struct {
	TokenID string `in:"path=token_id"`
}

type GetTokenResponse struct {
	ID           uuid.UUID        `json:"id"`
	TokenID      *string          `json:"tokenId,omitempty"`
	Type         models.TokenType `json:"type"`
	Name         string           `json:"name"`
	Symbol       string           `json:"symbol"`
	IsActive     bool             `json:"isActive"`
	BlockchainID uuid.UUID        `json:"blockchainId"`
	LogoURL      string           `json:"logoUrl"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type GetTokensResponse struct {
	Tokens []GetTokenResponse `json:"tokens"`
}

type GetMainWalletAddressRequest struct {
}

type GetMainWalletAddressResponse struct {
	Address string `json:"address"`
}
