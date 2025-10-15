package utils

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/client/grpc/utils"
	"fiatless/pkg/tron/wallet"
	"fmt"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

const (
	SUN_TO_TRX = 1000000 // 1 TRX = 1,000,000 SUN
)

type BaseWithdrawProcessor struct {
	PrivateKey   *ecdsa.PrivateKey
	OwnerAddress *address.TronAddress
	Client       api_tronpb.WalletClient
}

func NewBaseWithdrawProcessor(privateKey *ecdsa.PrivateKey, client api_tronpb.WalletClient) *BaseWithdrawProcessor {
	fromWallet := wallet.NewBaseTronWallet(privateKey)
	fromAddress := fromWallet.GetAddress()

	return &BaseWithdrawProcessor{
		PrivateKey:   privateKey,
		OwnerAddress: fromAddress,
		Client:       client,
	}
}

func (w *BaseWithdrawProcessor) GetAccountResource(ctx context.Context, addr *address.TronAddress) (*api_tronpb.AccountResourceMessage, error) {
	return w.Client.GetAccountResource(ctx, &core_tronpb.Account{Address: addr.Bytes()})
}

func (w *BaseWithdrawProcessor) GetAccount(ctx context.Context, addr *address.TronAddress) (*core_tronpb.Account, error) {
	return w.Client.GetAccount(ctx, &core_tronpb.Account{Address: addr.Bytes()})
}

func (w *BaseWithdrawProcessor) GetAccountTRXBalance(ctx context.Context, addr *address.TronAddress) (decimal.Decimal, error) {
	account, err := w.GetAccount(ctx, addr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get account: %v", err)
	}
	balance := decimal.NewFromInt(account.Balance).Div(decimal.NewFromInt(SUN_TO_TRX))
	return balance, nil
}

func (w *BaseWithdrawProcessor) BroadcastTransaction(ctx context.Context, tx *core_tronpb.Transaction) error {
	broadcastResult, err := w.Client.BroadcastTransaction(ctx, tx)
	if err != nil {
		txID, _ := GetTxID(tx)
		if broadcastResult != nil {
			return fmt.Errorf("failed to broadcast transaction (txID: %s): %v, result code: %s, message: %s",
				txID, err, broadcastResult.GetCode().String(), string(broadcastResult.GetMessage()))
		} else {
			return fmt.Errorf("failed to broadcast transaction (txID: %s): %v", txID, err)
		}
	}

	if !broadcastResult.GetResult() {
		txID, _ := GetTxID(tx)
		return fmt.Errorf("broadcast successful but transaction failed (txID: %s): result code: %s, message: %s",
			txID, broadcastResult.GetCode().String(), string(broadcastResult.GetMessage()))
	}

	return nil
}

// CalculateResources calculates all required resources (bandwidth, energy, activation fees) in a single call
func (w *BaseWithdrawProcessor) CalculateResources(ctx context.Context, chainParams *utils.ChainParams, fromAddr *address.TronAddress, toAddr *address.TronAddress, requiredBandwidth int64, requiredEnergy int64) (bandwidthCost decimal.Decimal, energyCost decimal.Decimal, activationFee decimal.Decimal, err error) {
	// Get resource information once
	fromResource, err := w.GetAccountResource(ctx, fromAddr)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("failed to get account resource: %v", err)
	}

	// Calculate bandwidth cost
	availableBandwidth := fromResource.FreeNetLimit + fromResource.NetLimit - fromResource.FreeNetUsed - fromResource.NetUsed
	if availableBandwidth < requiredBandwidth {
		bandwidthPrice, err := chainParams.GetChainParameter(ctx, "getTransactionFee")
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("failed to get bandwidth price: %v", err)
		}
		bandwidthCost = decimal.NewFromInt(requiredBandwidth * bandwidthPrice).Div(decimal.NewFromInt(SUN_TO_TRX))
	}

	// Calculate energy cost
	availableEnergy := fromResource.EnergyLimit - fromResource.EnergyUsed
	if availableEnergy < requiredEnergy {
		energyPrice, err := chainParams.GetChainParameter(ctx, "getEnergyFee")
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("failed to get energy price: %v", err)
		}
		energyNeeded := requiredEnergy - availableEnergy
		energyCost = decimal.NewFromInt(energyNeeded * energyPrice).Div(decimal.NewFromInt(SUN_TO_TRX))
	}

	// Calculate activation fee if needed
	if toAddr != nil {
		isActive := fromResource.FreeNetLimit != 0

		if !isActive {
			createAccountCost, err := chainParams.GetChainParameter(ctx, "getCreateNewAccountFeeInSystemContract")
			if err != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("failed to get create account fee: %v", err)
			}
			activationFee = decimal.NewFromInt(createAccountCost).Div(decimal.NewFromInt(SUN_TO_TRX))

			if availableBandwidth < requiredBandwidth {
				createAccountFee, err := chainParams.GetChainParameter(ctx, "getCreateAccountFee")
				if err != nil {
					return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("failed to get create account fee: %v", err)
				}
				bandwidthFee := decimal.NewFromInt(createAccountFee).Div(decimal.NewFromInt(SUN_TO_TRX))
				activationFee = activationFee.Add(bandwidthFee)
			}
		}
	}

	return bandwidthCost, energyCost, activationFee, nil
}
func GetTxID(tx *core_tronpb.Transaction) (string, error) {
	rawData, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction raw data: %v", err)
	}
	txHash := sha256.Sum256(rawData)
	txID := hex.EncodeToString(txHash[:])
	return txID, nil
}
