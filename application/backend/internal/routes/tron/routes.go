package tron

import (
	"fiatless/internal/constants"
	"fiatless/internal/models"
	"fiatless/internal/routes"
	"fiatless/internal/vars"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ggicci/httpin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func CreateWallet(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*CreateWalletRequest)

	wallet, err := vars.TronWalletManager.CreateWallet()
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

	if !request.IncludeResources && !request.IncludeTRC20 && !request.IncludeTRC10 && !request.IncludeTRX {
		utils.RespondWithError(w, http.StatusBadRequest, "No include parameters provided")
		return
	}
	var trc20Tokens []address.TronAddress
	var trc10Tokens []string

	var blockchainID uuid.UUID

	if request.IncludeTRC20 || request.IncludeTRC10 {
		blockchain, err := vars.Repositories.Blockchain.GetBlockchainBySymbolAndNetwork(r.Context(), "TRX", models.BlockchainNetwork(vars.Config.TronNetwork))
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		blockchainID = blockchain.ID
	}

	if request.IncludeTRC20 {
		tokens, err := vars.Repositories.Token.GetTokensByTypeAndBlockchain(r.Context(), models.TokenTypeTRC20, blockchainID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, token := range tokens {
			tronAddress, err := address.NewTronAddressFromBase58(*token.TokenID)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			trc20Tokens = append(trc20Tokens, *tronAddress)
		}
	}

	if request.IncludeTRC10 {
		tokens, err := vars.Repositories.Token.GetTokensByTypeAndBlockchain(r.Context(), models.TokenTypeTRC10, blockchainID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, token := range tokens {
			trc10Tokens = append(trc10Tokens, *token.TokenID)
		}
	}

	tronAddress, err := address.NewTronAddressFromBase58(request.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wallet, err := vars.TronWalletManager.GetWallet(*tronAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	balance, err := wallet.GetBalance(request.IncludeTRX, request.IncludeResources, trc20Tokens, trc10Tokens)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, balance)
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

	wallet, err := vars.TronWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tronAddress, err := address.NewTronAddressFromBase58(request.Payload.Address)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var tokenAddress *address.TronAddress
	if request.Payload.Token != "" {
		tokenAddress, err = address.NewTronAddressFromBase58(request.Payload.Token)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	tx, err := wallet.Withdraw(tronAddress, amount, tokenAddress)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var tokenID string
	var token *models.Token
	if tokenAddress != nil {
		tokenID = tokenAddress.String()
		token, err = vars.Repositories.Token.GetTokenByTokenID(r.Context(), &tokenID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else { // TRX
		tokenID = constants.GetTRXTokenID(vars.Config.TronNetwork)
		token, err = vars.Repositories.Token.GetToken(r.Context(), &models.Token{ID: uuid.MustParse(tokenID)})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	vars.Repositories.Transaction.CreateTransaction(r.Context(), &models.Transaction{
		ID:        uuid.New(),
		TxID:      tx.TransactionID,
		Fee:       tx.Fee,
		TokenID:   token.ID,
		ToAddress: tronAddress.String(),
		Amount:    amount,
		Type:      models.TransactionTypeWithdraw,
	})

	blockchainID := constants.GetTronBlockchainID(vars.Config.TronNetwork)
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

func StartSweeping(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*StartSweepingRequest)

	wallet, err := vars.TronWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Println(request.Payload.TokenID)
	token, err := vars.Repositories.Token.GetToken(r.Context(), &models.Token{ID: uuid.MustParse(request.Payload.TokenID)})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	walletDb, err := vars.Repositories.Wallet.GetWalletByAddress(r.Context(), wallet.GetAddress().String())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sweepingSession, err := vars.Repositories.SweepingSession.CreateSweepingSession(r.Context(), &models.SweepingSession{
		ID:                 uuid.New(),
		WalletMetaID:       walletDb.MetaID,
		TokenID:            token.ID,
		MinAmountThreshold: *request.Payload.MinAmount,
		CreatedAt:          time.Now(),
		Status:             models.SweepingSessionStatusPending,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wallets, err := vars.TronWalletManager.GetWallets(true)
	log.Println(wallets[0].GetAddress().String())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	mainWallet, err := vars.TronWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	vars.TaskManager.Add(vars.TaskFactory.CreateTronSweepingTask(sweepingSession, wallets, *mainWallet, *mainWallet, 10*time.Second, vars.Repositories, uuid.MustParse(constants.GetTRXTokenID(vars.Config.TronNetwork))))

	utils.RespondWithJSON(w, http.StatusOK, StartSweepingResponse{
		SessionID: sweepingSession.ID.String(),
	})
}

func GetMainWalletAddress(w http.ResponseWriter, r *http.Request) {
	mainWallet, err := vars.TronWalletManager.GetMainWallet()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, routes.GetMainWalletAddressResponse{Address: mainWallet.GetAddress().String()})
}

func GetTronTokens(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value(httpin.Input).(*routes.GetTokensRequest)

	tokens, err := vars.Repositories.Token.GetTokens(r.Context(), &models.Token{IsActive: request.IncludeInactive, BlockchainID: constants.GetTronBlockchainID(vars.Config.TronNetwork)})
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
