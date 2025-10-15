package bsc

import (
	"crypto/ecdsa"
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/bsc/address"
	"log"

	"fmt"

	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
)

type Service struct {
	client client.BlockchainClient
}

func NewService(client client.BlockchainClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetServiceName() string {
	return "bsc"
}

func (s *Service) GetBalance(address address.BSCAddress, includeBNB bool, bep20Tokens []address.BSCAddress) (*models.BSCWalletBalance, error) {
	requestParams := bp_models.BSCWalletBalanceRequest{
		Address:     address,
		IncludeBNB:  includeBNB,
		BEP20Tokens: bep20Tokens,
	}

	response, err := s.client.ExecuteCommand("/bsc/wallet/balance", requestParams)
	if err != nil {
		return nil, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return nil, fmt.Errorf("%s", response["error"].(string))
		}
		return nil, fmt.Errorf("invalid response format, result not found: %v", response)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var balance models.BSCWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(privateKey *ecdsa.PrivateKey, toAddress *address.BSCAddress, amount decimal.Decimal, tokenAddress *address.BSCAddress) (models.BSCWithdrawResponse, error) {
	privBytes := privateKey.D.Bytes()
	log.Printf("Withdrawing from address: %v", toAddress.String())
	privBytes = append(make([]byte, 32-len(privBytes)), privBytes...)
	privBase58 := base58.Encode(privBytes)

	requestParams := models.BSCWithdrawRequest{
		ContractAddress: tokenAddress,
		PrivateKey:      privBase58,
		ToAddress:       *toAddress,
		Amount:          amount,
	}

	response, err := s.client.ExecuteCommand("/bsc/wallet/withdraw", requestParams)
	if err != nil {
		return models.BSCWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return models.BSCWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return models.BSCWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return models.BSCWithdrawResponse{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var withdrawResponse models.BSCWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return models.BSCWithdrawResponse{}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return withdrawResponse, nil
}
