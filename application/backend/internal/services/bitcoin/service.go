package bitcoin

import (
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/bitcoin/address"
	"fmt"

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

func (s *Service) GetBalance(address address.BitcoinAddress) (*models.BitcoinWalletBalance, error) {
	requestParams := bp_models.BitcoinWalletBalanceRequest{
		Address: address,
	}

	response, err := s.client.ExecuteCommand("/bitcoin/wallet/balance", requestParams)
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

	var balance models.BitcoinWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(toAddress *address.BitcoinAddress, amount decimal.Decimal, utxos []bp_models.UTXO) (bp_models.BitcoinWithdrawResponse, error) {
	requestParams := bp_models.BitcoinWithdrawRequest{
		ToAddress: *toAddress,
		Amount:    amount,
		UTXOs:     utxos,
	}

	response, err := s.client.ExecuteCommand("/bitcoin/wallet/withdraw", requestParams)
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.BitcoinWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return bp_models.BitcoinWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var withdrawResponse bp_models.BitcoinWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return bp_models.BitcoinWithdrawResponse{}, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return withdrawResponse, nil
}

func (s *Service) ParseBlocks(walletAddresses []address.BitcoinAddress, utxos map[string]struct{}, latestBlockNumber *uint64) (bp_models.BitcoinParseBlocksResponse, error) {
	requestParams := bp_models.BitcoinParseBlocksRequest{
		WalletAddresses:   walletAddresses,
		UTXOs:             utxos,
		LatestBlockNumber: latestBlockNumber,
	}

	response, err := s.client.ExecuteCommand("/bitcoin/block/parse", requestParams)
	if err != nil {
		return bp_models.BitcoinParseBlocksResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		return bp_models.BitcoinParseBlocksResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.BitcoinParseBlocksResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var parseBlocks bp_models.BitcoinParseBlocksResponse
	if err := json.Unmarshal(jsonData, &parseBlocks); err != nil {
		return bp_models.BitcoinParseBlocksResponse{}, fmt.Errorf("failed to unmarshal result: %v", err)
	}

	return parseBlocks, nil
}
