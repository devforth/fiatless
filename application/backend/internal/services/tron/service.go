package tron

import (
	"crypto/ecdsa"
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/tron/address"
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

func (s *Service) GetBalance(address address.TronAddress, includeTRX, includeResources bool, trc20Tokens []address.TronAddress, trc10Tokens []string) (*models.TronWalletBalance, error) {
	params := bp_models.TronWalletBalanceRequest{
		Address:          address,
		IncludeTRX:       includeTRX,
		IncludeResources: includeResources,
		TRC20Tokens:      trc20Tokens,
		TRC10Tokens:      trc10Tokens,
	}

	response, err := s.client.ExecuteCommand("/tron/wallet/balance", params)
	if err != nil {
		return nil, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return nil, fmt.Errorf("%s", response["error"].(string))
		}
		return nil, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %v", err)
	}

	var balance models.TronWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result into models.TronWalletBalance: %v", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(privateKey *ecdsa.PrivateKey, toAddress address.TronAddress, amount decimal.Decimal, token *address.TronAddress) (bp_models.TronWithdrawResponse, error) {
	privBytes := privateKey.D.Bytes()
	privBytes = append(make([]byte, 32-len(privBytes)), privBytes...)
	privBase58 := base58.Encode(privBytes)

	params := bp_models.TronWithdrawRequest{
		ContractAddress: token,
		PrivateKey:      privBase58,
		ToAddress:       toAddress,
		Amount:          amount,
		FeeLimit:        10_000_000,
	}

	response, err := s.client.ExecuteCommand("/tron/wallet/withdraw", params)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.TronWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var withdrawResponse bp_models.TronWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to unmarshal result into models.WithdrawResponse: %v", err)
	}

	return withdrawResponse, nil
}

func (s *Service) ParseBlocks(walletAddresses []address.TronAddress, tokens []bp_models.ParseBlocksToken, latestBlockNumber *uint64) (bp_models.ParseBlocksResponse, error) {
	params := bp_models.ParseBlocksRequest{
		WalletAddresses:   walletAddresses,
		LatestBlockNumber: latestBlockNumber,
		Tokens:            tokens,
	}

	response, err := s.client.ExecuteCommand("/tron/block/parse", params)
	if err != nil {
		return bp_models.ParseBlocksResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		return bp_models.ParseBlocksResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.ParseBlocksResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var parseBlocks bp_models.ParseBlocksResponse
	if err := json.Unmarshal(jsonData, &parseBlocks); err != nil {
		return bp_models.ParseBlocksResponse{}, fmt.Errorf("failed to unmarshal result into models.ParseBlocks: %v", err)
	}

	return parseBlocks, nil
}
