package xrpl

import (
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/utils"
	"net/http"
	"strings"

	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/ggicci/httpin"
	"github.com/shopspring/decimal"
)

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.XRPLWalletManager.CreateWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := CreateWalletResponse{
		Address: wallet.GetAddress(),
		TagID:   request.Payload.TagID,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*GetBalanceRequest)

	wallet, err := vars.XRPLWalletManager.GetWallet(xrpl_types.Address(request.Address))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	balance, err := wallet.GetBalance()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, GetBalanceResponse{Balance: balance.Balance})
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*WithdrawRequest)

	address := xrpl_types.Address(request.Payload.Address)

	wallet, err := vars.XRPLWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	amount, err := decimal.NewFromString(request.Payload.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := wallet.Withdraw(address, amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	blockchainID := constants.GetXRPLBlockchainID(vars.Config.XRPLNetwork)
	blockchain, err := vars.Repositories.Blockchain.GetBlockchain(r.Context(), &models.Blockchain{ID: blockchainID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, WithdrawResponse{TransactionID: tx.TransactionID, ExplorerURL: strings.Replace(blockchain.ExplorerURL, "{{tx_id}}", tx.TransactionID, 1)})
}

func GetMainWalletAddress(w http.ResponseWriter, r *http.Request) {
	mainWallet, err := vars.XRPLWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}

func GetXRPLTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetXRPLBlockchainID(vars.Config.XRPLNetwork)})
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
