package ethereum

import (
	"context"
	"errors"
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/ethereum/address"
	"fiatless/pkg/utils"
	"log"
	"net/http"
	"strings"

	"github.com/ggicci/httpin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func getBalance(addrs *address.EthereumAddress, includeERC20 bool, includeETH bool) (*GetBalanceResponse, error) {
	wallet, err := vars.EthereumWalletManager.GetWallet(*addrs)
	if err != nil {
		return nil, err
	}

	if !includeETH && !includeERC20 {
		return nil, errors.New("no include parameters provided")
	}

	blockchainID := constants.GetEthereumBlockchainID(vars.Config.EthereumNetwork)
	tokens, err := vars.Repositories.Token.GetTokensByTypeAndBlockchain(context.Background(), models.TokenTypeERC20, blockchainID)
	if err != nil {
		return nil, err
	}

	response := GetBalanceResponse{}
	erc20Tokens := []address.EthereumAddress{}
	tokensMap := make(map[string]models.Token)
	if includeERC20 {
		for _, token := range tokens {
			erc20Token, err := address.NewEthereumAddressFromString(*token.TokenID)
			if err != nil {
				return nil, err
			}
			erc20Tokens = append(erc20Tokens, *erc20Token)
			tokensMap[erc20Token.String()] = token
		}
	}

	balance, err := wallet.GetBalance(includeETH, erc20Tokens)
	if err != nil {
		return nil, err
	}

	response.Balance = &balance.ETHBalance
	for _, token := range balance.ERC20Balances {
		response.ERC20Balances = append(response.ERC20Balances, ERC20Balance{
			TokenID: tokensMap[token.ContractAddress.String()].ID,
			Balance: token.Balance,
		})
	}

	return &response, nil
}

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.EthereumWalletManager.CreateWallet()
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

	ethereumAddress, err := address.NewEthereumAddressFromString(request.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := getBalance(ethereumAddress, request.IncludeERC20, request.IncludeETH)
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
	ethereumAddress, err := address.NewEthereumAddressFromString(request.Payload.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	mainWallet, err := vars.EthereumWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if mainWallet.GetAddress().String() == ethereumAddress.String() {
		utils.RespondWithError(w, http.StatusBadRequest, "Cannot withdraw to main wallet")
		return
	}

	amount, err := decimal.NewFromString(request.Payload.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var tokenAddress *address.EthereumAddress
	log.Println("token", token.TokenID, token.Type)
	if token.TokenID != nil && token.Type == models.TokenTypeERC20 {
		tokenAddress, err = address.NewEthereumAddressFromString(*token.TokenID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else if token.TokenID == nil && token.Type == models.TokenTypeNative {
		tokenAddress = nil
	} else {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid token")
		return
	}

	tx, err := mainWallet.Withdraw(ethereumAddress, amount, tokenAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vars.Repositories.Transaction.CreateTransaction(r.Context(), &models.Transaction{
		ID:        uuid.New(),
		TxID:      tx.TransactionID,
		Fee:       decimal.Zero,
		TokenID:   token.ID,
		ToAddress: ethereumAddress.String(),
		Amount:    amount,
		Type:      models.TransactionTypeWithdraw,
	})
	blockchainID := constants.GetEthereumBlockchainID(vars.Config.EthereumNetwork)
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

func GetEthereumTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetEthereumBlockchainID(vars.Config.EthereumNetwork)})
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
	mainWallet, err := vars.EthereumWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}
