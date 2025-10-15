package xrpl

import (
	"encoding/json"
	"fiatless/internal/blockchain/client"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fmt"

	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
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

func (s *Service) GetBalance(address xrpl_types.Address) (*models.XRPLWalletBalance, error) {
	params := bp_models.XRPLWalletBalanceRequest{
		Address: address,
	}

	response, err := s.client.ExecuteCommand("/xrpl/wallet/balance", params)
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

	var balance models.XRPLWalletBalance
	if err := json.Unmarshal(jsonData, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result into models.XRPLWalletBalance: %v", err)
	}

	return &balance, nil
}

func (s *Service) Withdraw(privateKey string, toAddress xrpl_types.Address, amount decimal.Decimal) (bp_models.XRPLWithdrawResponse, error) {
	params := bp_models.XRPLWithdrawRequest{
		PrivateKey: privateKey,
		ToAddress:  toAddress,
		Amount:     amount,
	}

	response, err := s.client.ExecuteCommand("/xrpl/wallet/withdraw", params)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, err
	}

	result, ok := response["result"]
	if !ok {
		if response["error"] != nil {
			return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("%s", response["error"].(string))
		}
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("invalid response format, result not found")
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to marshal result: %v", err)
	}

	var withdrawResponse bp_models.XRPLWithdrawResponse
	if err := json.Unmarshal(jsonData, &withdrawResponse); err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to unmarshal result into models.WithdrawResponse: %v", err)
	}

	return withdrawResponse, nil
}
