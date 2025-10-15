package bitcoin

import (
	"context"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/bitcoin/address"
	"fiatless/pkg/utils"
	"log"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/shopspring/decimal"
)

func getBalance(ctx context.Context, address *address.BitcoinAddress) (*decimal.Decimal, error) {
	balance, err := vars.Repositories.UTXO.GetBalance(ctx, address.String())
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*GetBalanceRequest)

	bitcoinAddress, err := address.NewBitcoinAddressFromString(request.Address, vars.Config.BitcoinNetwork)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	balance, err := getBalance(r.Context(), bitcoinAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, GetBalanceResponse{
		Balance: balance,
	})
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.BitcoinWalletManager.CreateWallet()
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

func Withdraw(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*WithdrawRequest)

	addresses, err := vars.BitcoinWalletManager.GetAllAddresses()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var addressesStrings []string
	for _, address := range addresses {
		addressesStrings = append(addressesStrings, address.String())
	}
	log.Println("addressesStrings", addressesStrings)

	result, err := vars.BitcoinWalletManager.Withdraw(&request.Payload.Address, request.Payload.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	blockchainID := constants.GetBitcoinBlockchainID(vars.Config.BitcoinNetwork)
	blockchain, err := vars.Repositories.Blockchain.GetBlockchain(r.Context(), &models.Blockchain{ID: blockchainID})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := WithdrawResponse{
		TransactionID: result.TransactionID,
		ExplorerURL:   strings.Replace(blockchain.ExplorerURL, "{{tx_id}}", result.TransactionID, 1),
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func GetMainWalletAddress(w http.ResponseWriter, r *http.Request) {
	mainWallet, err := vars.BitcoinWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}

func GetBitcoinTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetBitcoinBlockchainID(vars.Config.BitcoinNetwork)})
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
