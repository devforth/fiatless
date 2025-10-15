package routes

import (
	"fiatless/internal/models"
	"fiatless/internal/vars"
	"fiatless/pkg/utils"
	"net/http"

	"fiatless/internal/constants"

	"github.com/ggicci/httpin"
	"github.com/google/uuid"
)

func GetTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*GetTokensRequest)
	token := &models.Token{}
	if !request.IncludeInactive {
		token.IsActive = true
	}
	tokens, err := vars.Repositories.Token.GetTokensByBlockchains(r.Context(), token, []uuid.UUID{
		constants.GetEthereumBlockchainID(vars.Config.EthereumNetwork),
		constants.GetTronBlockchainID(vars.Config.TronNetwork),
		constants.GetBSCBlockchainID(vars.Config.BSCNetwork),
		constants.GetBitcoinBlockchainID(vars.Config.BitcoinNetwork),
		constants.GetXRPLBlockchainID(vars.Config.XRPLNetwork),
		constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := GetTokensResponse{}

	for _, token := range tokens {
		response.Tokens = append(response.Tokens, GetTokenResponse{
			ID:           token.ID,
			TokenID:      token.TokenID,
			Type:         token.Type,
			Name:         token.Name,
			Symbol:       token.Symbol,
			IsActive:     token.IsActive,
			BlockchainID: token.BlockchainID,
			LogoURL:      token.LogoURL,
			CreatedAt:    token.CreatedAt,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*GetTokenRequest)
	tokenID, err := uuid.Parse(request.TokenID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid token ID")
		return
	}

	token, err := vars.Repositories.Token.GetTokensByBlockchains(r.Context(), &models.Token{
		ID: tokenID,
	}, []uuid.UUID{
		constants.GetEthereumBlockchainID(vars.Config.EthereumNetwork),
		constants.GetTronBlockchainID(vars.Config.TronNetwork),
		constants.GetBSCBlockchainID(vars.Config.BSCNetwork),
		constants.GetBitcoinBlockchainID(vars.Config.BitcoinNetwork),
		constants.GetXRPLBlockchainID(vars.Config.XRPLNetwork),
		constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(token) == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Token not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, GetTokenResponse{
		ID:           token[0].ID,
		TokenID:      token[0].TokenID,
		Type:         token[0].Type,
		Name:         token[0].Name,
		Symbol:       token[0].Symbol,
		IsActive:     token[0].IsActive,
		BlockchainID: token[0].BlockchainID,
		LogoURL:      token[0].LogoURL,
		CreatedAt:    token[0].CreatedAt,
	})
}
