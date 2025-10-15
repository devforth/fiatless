package solana

import (
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/solana/address"
	"fiatless/pkg/utils"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.SolanaWalletManager.CreateWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := CreateWalletResponse{
		Address: *wallet.GetAddress(),
		TagID:   request.Payload.TagID,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*GetBalanceRequest)

	solanaAddress, err := address.NewSolanaAddressFromString(request.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	wallet, err := vars.SolanaWalletManager.GetWallet(*solanaAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	balance, err := wallet.GetBalance()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, GetBalanceResponse{Balance: &balance.SOLBalance})
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*WithdrawRequest)

	amount, err := decimal.NewFromString(request.Payload.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if amount.LessThan(decimal.NewFromInt(0)) {
		utils.RespondWithError(w, http.StatusBadRequest, "Amount must be greater than 0")
		return
	}

	wallet, err := vars.SolanaWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	toAddress, err := address.NewSolanaAddressFromString(request.Payload.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := wallet.Withdraw(toAddress, amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	blockchainID := constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork)
	blockchain, err := vars.Repositories.Blockchain.GetBlockchain(r.Context(), &models.Blockchain{ID: blockchainID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	solanaToken := constants.GetSolanaTokenID(vars.Config.SolanaNetwork)
	vars.Repositories.Transaction.CreateTransaction(r.Context(), &models.Transaction{
		ID:        uuid.New(),
		TxID:      tx.TransactionID,
		Fee:       decimal.Zero,
		TokenID:   uuid.MustParse(solanaToken),
		ToAddress: toAddress.String(),
		Amount:    amount,
		Type:      models.TransactionTypeWithdraw,
	})

	response := WithdrawResponse{
		TransactionID: tx.TransactionID,
		ExplorerURL:   strings.Replace(blockchain.ExplorerURL, "{{tx_id}}", tx.TransactionID, 1),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetMainWalletAddress(w http.ResponseWriter, r *http.Request) {
	mainWallet, err := vars.SolanaWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}

func GetSolanaTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork)})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := routes.GetTokensResponse{}

	for _, token := range tokens {
		response.Tokens = append(response.Tokens, routes.GetTokenResponse{
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
