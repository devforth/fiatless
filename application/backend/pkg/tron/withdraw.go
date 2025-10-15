package tron

import (
	"context"
	"crypto/ecdsa"
	bp_models "fiatless/internal/bp-models"
	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/client/grpc/utils"
	tron_utils "fiatless/pkg/tron/utils"
	common_utils "fiatless/pkg/utils"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/holiman/uint256"
	"github.com/shopspring/decimal"
)

type WithdrawProcessor struct {
	tron_utils.BaseWithdrawProcessor
}

func NewWithdrawProcessor(client api_tronpb.WalletClient, privKey *ecdsa.PrivateKey) *WithdrawProcessor {
	from, _ := address.NewTronAddressFromPrivateKey(privKey)
	return &WithdrawProcessor{
		BaseWithdrawProcessor: tron_utils.BaseWithdrawProcessor{
			PrivateKey:   privKey,
			OwnerAddress: from,
			Client:       client,
		},
	}
}

func (w *WithdrawProcessor) EstimateEnergy(ctx context.Context, contractData *core_tronpb.TriggerSmartContract) (int64, error) {
	tx, err := w.Client.TriggerConstantContract(ctx, contractData)
	if err != nil {
		return 0, fmt.Errorf("failed to trigger constant contract: %v", err)
	}

	return tx.EnergyUsed, nil
}

func (w *WithdrawProcessor) DetermineFeeLimit(ctx context.Context, chainParams *utils.ChainParams, energyUsed int64) (int64, error) {
	energyPrice, err := chainParams.GetChainParameter(ctx, "getEnergyFee")
	if err != nil {
		return 0, fmt.Errorf("failed to get energy price: %v", err)
	}

	return energyUsed * energyPrice, nil
}

func (w *WithdrawProcessor) TransferTRC20Token(ctx context.Context, contractAddress *address.TronAddress, to *address.TronAddress, amount decimal.Decimal) (bp_models.TronWithdrawResponse, error) {
	// 1. Get token decimals and prepare amount
	var wg sync.WaitGroup
	var decimalsResult []any
	var decimalsErr error
	var balanceResult []any
	var balanceErr error

	contract, err := NewTRC20Contract(w.Client, contractAddress)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to create TRC20 contract: %v", err)
	}

	decimalsMethod, err := contract.Functions.Get("decimals")
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get decimals method: %v", err)
	}

	balanceMethod, err := contract.Functions.Get("balanceOf")
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get balance method: %v", err)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		decimalsResult, decimalsErr = decimalsMethod.Call(ctx)
	}()

	go func() {
		defer wg.Done()
		balanceResult, balanceErr = balanceMethod.Call(ctx, w.OwnerAddress.Bytes())
	}()

	wg.Wait()

	if decimalsErr != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get decimals: %v", decimalsErr)
	}

	decimals := decimalsResult[0].(uint8)

	if balanceErr != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get balance: %v", balanceErr)
	}
	// 2. Convert balance and amount to decimals
	balanceWithDecimals := uint256.NewInt(0).SetBytes(balanceResult[0].([]byte))
	amountWithDecimals := uint256.NewInt(0).SetBytes(amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals)))).BigInt().Bytes())

	balanceDec, err := decimal.NewFromString(balanceWithDecimals.String())
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to convert balance to decimal: %v", err)
	}
	balance := balanceDec.Div(decimal.NewFromInt(int64(decimals)))

	if amount.GreaterThan(balance) {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("insufficient token balance. Need %s tokens, have %s tokens",
			amount.String(), balance.String())
	}

	// 3. Prepare contract call data
	method := "transfer"
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	arguments := abi.Arguments{
		{Type: addressType},
		{Type: uint256Type},
	}

	data, err := arguments.Pack(to, amountWithDecimals)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to pack function data for transfer: %v", err)
	}

	selector := common_utils.GetFunctionSelector(method + "(address,uint256)")
	callData := append(selector, data...)

	// 4. Build transaction
	builder := tron_utils.NewTronTransactionBuilder(w.PrivateKey)
	block, err := w.Client.GetNowBlock2(ctx, &api_tronpb.EmptyMessage{})
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get current block: %v", err)
	}
	builder.SetReferenceBlock(block)

	contractData := core_tronpb.TriggerSmartContract{
		OwnerAddress:    w.OwnerAddress.Bytes(),
		ContractAddress: contractAddress.Bytes(),
		Data:            callData,
		CallValue:       0,
	}

	builder.SetContractType(core_tronpb.Transaction_Contract_TriggerSmartContract)
	builder.SetContractData(&contractData)

	// 5. Estimate energy and set fee limit
	estimateEnergy, err := w.EstimateEnergy(ctx, &contractData)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to estimate energy: %v", err)
	}

	chainParams := utils.NewChainParams(w.Client)
	err = chainParams.InitChainParams(ctx)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to initialize chain params: %v", err)
	}

	feeLimit, err := w.DetermineFeeLimit(ctx, chainParams, estimateEnergy)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to determine fee limit: %v", err)
	}

	builder.SetFeeLimit(feeLimit)

	// 6. Build and sign transaction
	tx, err := builder.Build()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to build transaction: %v", err)
	}

	_, err = builder.Sign()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 7. Calculate required resources
	estimateBandwidth, err := builder.CalculateBandwidth()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to calculate bandwidth: %v", err)
	}

	bandwidthCost, energyCost, _, err := w.CalculateResources(ctx, chainParams, w.OwnerAddress, nil, estimateBandwidth, estimateEnergy)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to calculate resources: %v", err)
	}

	requiredTRX := bandwidthCost.Add(energyCost)

	// 8. Check TRX balance is sufficient for fees
	trxBalance, err := w.GetAccountTRXBalance(ctx, w.OwnerAddress)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get TRX balance: %v", err)
	}

	if requiredTRX.GreaterThan(trxBalance) {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("insufficient TRX balance for fees. Need %s TRX, have %s TRX",
			requiredTRX.String(), trxBalance.String())
	}

	// 9. Broadcast transaction
	err = w.BroadcastTransaction(ctx, tx)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	// 10. Get transaction ID
	txID, err := tron_utils.GetTxID(tx)
	if err != nil {
		fmt.Printf("Warning: transaction broadcasted but failed to get txID: %v\n", err)
		return bp_models.TronWithdrawResponse{TransactionID: "(unknown, check logs)"}, nil
	}

	return bp_models.TronWithdrawResponse{
		TransactionID: txID,
		Fee:           decimal.NewFromInt(tx.Ret[0].GetFee()).Div(decimal.NewFromInt(1000000)), //The total number of TRX burned in this transaction, including TRX burned for bandwidth/energy, memo fee, account activation fee, multi-signature fee and other fees
	}, nil
}

func (w *WithdrawProcessor) Withdraw(ctx context.Context, toAddress *address.TronAddress, amount decimal.Decimal) (bp_models.TronWithdrawResponse, error) {
	// 1. Build transaction
	builder := tron_utils.NewTronTransactionBuilder(w.PrivateKey)
	block, err := w.Client.GetNowBlock2(ctx, nil)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get block: %v", err)
	}

	amountBigInt := amount.Mul(decimal.NewFromInt(1000000)).BigInt()

	builder.SetReferenceBlock(block)
	tx, err := builder.
		TransferTRX(toAddress, amountBigInt).
		Build()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to create transaction: %v", err)
	}

	_, err = builder.Sign()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// 2. Calculate required bandwidth
	txBandwidth, err := builder.CalculateBandwidth()
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to calculate bandwidth: %v", err)
	}

	// 3. Calculate fees

	chainParams := utils.NewChainParams(w.Client)
	err = chainParams.InitChainParams(ctx)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to initialize chain params: %v", err)
	}

	bandwidthCost, _, activationFee, err := w.CalculateResources(ctx, chainParams, w.OwnerAddress, toAddress, txBandwidth, 0)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to calculate resources: %v", err)
	}

	// 4. Calculate total cost
	totalCost := amount.Add(activationFee).Add(bandwidthCost)

	// 5. Check if balance is sufficient
	accountBalance, err := w.GetAccountTRXBalance(ctx, w.OwnerAddress)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get account balance: %v", err)
	}

	if totalCost.GreaterThan(accountBalance) {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("insufficient balance. Need %s TRX, have %s TRX",
			totalCost.String(), accountBalance.String())
	}

	// 6. Broadcast transaction
	err = w.BroadcastTransaction(ctx, tx)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	// 7. Get transaction ID
	txID, err := tron_utils.GetTxID(tx)
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// 8. Get transaction
	txInfo, err := w.Client.GetTransactionInfoById(ctx, &api_tronpb.BytesMessage{Value: []byte(txID)})
	if err != nil {
		return bp_models.TronWithdrawResponse{}, fmt.Errorf("failed to get transaction: %v", err)
	}

	txFee := decimal.NewFromInt(txInfo.Fee).Div(decimal.NewFromInt(1000000))

	return bp_models.TronWithdrawResponse{
		TransactionID: txID,
		Fee:           txFee,
	}, nil
}
