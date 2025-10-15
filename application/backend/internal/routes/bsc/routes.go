package bsc

import (
	"context"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/bsc/address"
	"fiatless/pkg/utils"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func getBalance(addrs *address.BSCAddress, includeBNB bool, includeBEP20 bool) (*GetBalanceResponse, error) {
	wallet, err := vars.BSCWalletManager.GetWallet(*addrs)
	if err != nil {
		return nil, err
	}

	if !includeBNB && !includeBEP20 {
		return nil, errors.New("no include parameters provided")
	}

	blockchainID := constants.GetBSCBlockchainID(vars.Config.BSCNetwork)
	tokens, err := vars.Repositories.Token.GetTokensByTypeAndBlockchain(context.Background(), models.TokenTypeBEP20, blockchainID)
	if err != nil {
		return nil, err
	}

	response := GetBalanceResponse{}
	bep20Tokens := []address.BSCAddress{}
	tokensMap := make(map[string]models.Token)
	if includeBNB {
		for _, token := range tokens {
			bep20Token, err := address.NewBSCAddressFromString(*token.TokenID)
			if err != nil {
				return nil, err
			}
			bep20Tokens = append(bep20Tokens, *bep20Token)
			tokensMap[bep20Token.String()] = token
		}
	}

	balance, err := wallet.GetBalance(includeBNB, bep20Tokens)
	if err != nil {
		return nil, err
	}

	response.Balance = &balance.BNBBalance
	for _, token := range balance.BEP20Balances {
		response.BEP20Balances = append(response.BEP20Balances, BEP20Balance{
			TokenID: tokensMap[token.ContractAddress.String()].ID,
			Balance: token.Balance,
		})
	}

	return &response, nil
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.BSCWalletManager.CreateWallet()
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

	bscAddress, err := address.NewBSCAddressFromString(request.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := getBalance(bscAddress, request.IncludeBNB, request.IncludeBEP20)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*WithdrawRequest)

	var token *models.Token
	if request.Payload.TokenID != "" {
		id, err := uuid.Parse(request.Payload.TokenID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		token, err = vars.Repositories.Token.GetToken(r.Context(), &models.Token{ID: id})
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	bscAddress, err := address.NewBSCAddressFromString(request.Payload.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	mainWallet, err := vars.BSCWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if mainWallet.GetAddress().String() == bscAddress.String() {
		utils.RespondWithError(w, http.StatusBadRequest, "Cannot withdraw to main wallet")
		return
	}

	amount, err := decimal.NewFromString(request.Payload.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var tokenAddress *address.BSCAddress
	switch token.Type {
	case models.TokenTypeBEP20:
		tokenAddress, err = address.NewBSCAddressFromString(*token.TokenID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	case models.TokenTypeNative:
		tokenAddress = nil
	default:
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid token")
		return
	}

	tx, err := mainWallet.Withdraw(bscAddress, amount, tokenAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vars.Repositories.Transaction.CreateTransaction(r.Context(), &models.Transaction{
		ID:        uuid.New(),
		TxID:      tx.TransactionID,
		Fee:       decimal.Zero,
		TokenID:   token.ID,
		ToAddress: bscAddress.String(),
		Amount:    amount,
		Type:      models.TransactionTypeWithdraw,
	})
	blockchainID := constants.GetBSCBlockchainID(vars.Config.BSCNetwork)
	blockchain, err := vars.Repositories.Blockchain.GetBlockchain(r.Context(), &models.Blockchain{ID: blockchainID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := WithdrawResponse{
		TransactionID: tx.TransactionID,
		ExplorerURL:   strings.Replace(blockchain.ExplorerURL, "{{tx_id}}", tx.TransactionID, 1),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetBSCTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetBSCBlockchainID(vars.Config.BSCNetwork)})
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

func GetMainWalletAddress(w http.ResponseWriter, r *http.Request) {
	mainWallet, err := vars.BSCWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}
