package solana

import (
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/solana/address"
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

func (s *Service) GetBalance(address address.SolanaAddress) (*models.SolanaWalletBalance, error) {
	params := bp_models.SolanaWalletBalanceRequest{Address: address}

	response, err := s.client.ExecuteCommand("/solana/wallet/balance", params)
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

	var balance models.SolanaWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result into models.SolanaWalletBalance: %v", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(privateKey []byte, toAddress address.SolanaAddress, amount decimal.Decimal) (bp_models.SolanaWithdrawResponse, error) {
	privBase58 := base58.Encode(privateKey)

	params := bp_models.SolanaWithdrawRequest{ExpandedPrivateKey: privBase58, ToAddress: toAddress, Amount: amount}

	response, err := s.client.ExecuteCommand("/solana/wallet/withdraw", params)
	if err != nil {
		return bp_models.SolanaWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.SolanaWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return bp_models.SolanaWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.SolanaWithdrawResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var withdrawResponse bp_models.SolanaWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return bp_models.SolanaWithdrawResponse{}, fmt.Errorf("failed to unmarshal result into models.WithdrawResponse: %v", err)
	}

	return withdrawResponse, nil
}

func (s *Service) ParseBlocks(walletAddresses []address.SolanaAddress, tokenID string, latestBlockNumber *uint64) (bp_models.ParseBlocksResponse, error) {
	params := bp_models.SolanaParseBlocksRequest{
		WalletAddresses:   walletAddresses,
		LatestBlockNumber: latestBlockNumber,
		TokenID:           tokenID,
	}

	response, err := s.client.ExecuteCommand("/solana/block/parse", params)
	if err != nil {
		return bp_models.ParseBlocksResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.ParseBlocksResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
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

func (s *Service) GetWalletTransactions(walletAddress address.SolanaAddress, latestTransactionId *string) (bp_models.SolanaGetWalletTransactionsResponse, error) {
	params := bp_models.SolanaGetWalletTransactionsRequest{
		WalletAddress:       walletAddress,
		LatestTransactionId: latestTransactionId,
	}

	response, err := s.client.ExecuteCommand("/solana/wallet/transactions", params)
	if err != nil {
		return bp_models.SolanaGetWalletTransactionsResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		return bp_models.SolanaGetWalletTransactionsResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.SolanaGetWalletTransactionsResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var transactions bp_models.SolanaGetWalletTransactionsResponse
	if err := json.Unmarshal(jsonData, &transactions); err != nil {
		return bp_models.SolanaGetWalletTransactionsResponse{}, fmt.Errorf("failed to unmarshal result into models.SolanaTransaction: %v", err)
	}

	return transactions, nil
}
