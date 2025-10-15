package ethereum

import (
	"crypto/ecdsa"
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/ethereum/address"
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

func (s *Service) GetBalance(address address.EthereumAddress, includeETH bool, erc20Tokens []address.EthereumAddress) (*models.EthereumWalletBalance, error) {
	requestParams := bp_models.EthereumWalletBalanceRequest{
		Address:     address,
		IncludeETH:  includeETH,
		ERC20Tokens: erc20Tokens,
	}

	response, err := s.client.ExecuteCommand("/ethereum/wallet/balance", requestParams)
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

	var balance models.EthereumWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(privateKey *ecdsa.PrivateKey, toAddress *address.EthereumAddress, amount decimal.Decimal, tokenAddress *address.EthereumAddress) (bp_models.EthereumWithdrawResponse, error) {
	privBytes := privateKey.D.Bytes()
	log.Printf("Withdrawing from address: %v", toAddress.String())
	privBytes = append(make([]byte, 32-len(privBytes)), privBytes...)
	privBase58 := base58.Encode(privBytes)

	requestParams := bp_models.EthereumWithdrawRequest{
		ContractAddress: tokenAddress,
		PrivateKey:      privBase58,
		ToAddress:       *toAddress,
		Amount:          amount,
	}

	response, err := s.client.ExecuteCommand("/ethereum/wallet/withdraw", requestParams)
	if err != nil {
		return bp_models.EthereumWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var withdrawResponse bp_models.EthereumWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return withdrawResponse, nil
}
