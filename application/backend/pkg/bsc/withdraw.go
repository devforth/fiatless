package bsc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"sync"

	"fiatless/pkg/bsc/address"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

type WithdrawProcessor struct {
	Client  *ethclient.Client
	PrivKey *ecdsa.PrivateKey
	From    common.Address
}

func NewWithdrawProcessor(client *ethclient.Client, privKey *ecdsa.PrivateKey) *WithdrawProcessor {
	from, _ := address.NewBSCAddressFromPrivateKey(privKey)
	return &WithdrawProcessor{
		Client:  client,
		PrivKey: privKey,
		From:    common.HexToAddress(from.String()),
	}
}

func (w *WithdrawProcessor) TransferToken(ctx context.Context, contract *Contract, to common.Address, amount *decimal.Decimal) (*types.Transaction, error) {
	balanceOfFunc, err := contract.Functions.Get("balanceOf")
	if err != nil {
		return nil, fmt.Errorf("balanceOf method not found: %w", err)
	}

	decimalsFunc, err := contract.Functions.Get("decimals")
	if err != nil {
		return nil, fmt.Errorf("decimals method not found: %w", err)
	}

	transferFunc, err := contract.Functions.Get("transfer")
	if err != nil {
		return nil, fmt.Errorf("transfer method not found: %w", err)
	}

	var wg sync.WaitGroup
	var balanceOfCall []any
	var balanceEthWithDecimals *big.Int
	var gasPrice *big.Int
	var gasLimit uint64
	var nonce uint64
	var chainID *big.Int
	var decimalsOfCall []any

	var BalanceErr error
	var BalanceEthErr error
	var GasPriceErr error
	var GasLimitErr error
	var NonceErr error
	var ChainIDErr error
	var DecimalsErr error
	wg.Add(4)

	go func() {
		defer wg.Done()
		balanceEthWithDecimals, BalanceEthErr = w.Client.BalanceAt(ctx, w.From, nil)
	}()

	go func() {
		defer wg.Done()
		balanceOfCall, BalanceErr = balanceOfFunc.Call(ctx, w.From)
	}()

	go func() {
		defer wg.Done()
		decimalsOfCall, DecimalsErr = decimalsFunc.Call(ctx)
	}()

	go func() {
		defer wg.Done()
		gasPrice, GasPriceErr = w.Client.SuggestGasPrice(ctx)
	}()

	wg.Wait()

	if BalanceEthErr != nil {
		return nil, fmt.Errorf("failed to get balance: %v", BalanceEthErr)
	}

	if BalanceErr != nil {
		return nil, fmt.Errorf("failed to get balance: %v", BalanceErr)
	}
	if GasPriceErr != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", GasPriceErr)
	}

	if DecimalsErr != nil {
		return nil, fmt.Errorf("failed to get decimals: %v", DecimalsErr)
	}

	decimals := decimalsOfCall[0].(uint8)
	amountWithDecimals := amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals))))
	balanceWithDecimals := balanceOfCall[0].(*big.Int)
	balance := decimal.NewFromBigInt(balanceWithDecimals, -int32(decimals))

	if balance.Cmp(*amount) < 0 {
		return nil, fmt.Errorf("insufficient token balance. Need: %v, Have: %v", amount, balance)
	}
	data, err := transferFunc.ABIEntry.Inputs.Pack(to, amountWithDecimals.BigInt())
	if err != nil {
		return nil, err
	}
	txData := append(transferFunc.ABIEntry.ID, data...)

	gasLimit, GasLimitErr = w.Client.EstimateGas(ctx, ethereum.CallMsg{
		From:  w.From,
		To:    &contract.Address,
		Value: nil,
		Data:  txData,
	})
	if GasLimitErr != nil {
		re := regexp.MustCompile(`failed with (\d+) gas`)
		match := re.FindStringSubmatch(GasLimitErr.Error())
		if len(match) > 1 {
			gasLimit, _ = strconv.ParseUint(match[1], 10, 64)
		} else {
			return nil, fmt.Errorf("failed to estimate gas: %v", GasLimitErr)
		}
	}

	balanceEth := decimal.NewFromBigInt(balanceEthWithDecimals, int32(-18))
	finalAmountEth := decimal.NewFromBigInt(new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice), int32(-18))

	log.Printf("balanceEth: %v", balanceEth)
	log.Printf("finalAmountEth: %v", finalAmountEth)

	if balanceEth.Cmp(finalAmountEth) < 0 {
		return nil, fmt.Errorf("insufficient eth balance. Need: %v, Have: %v", finalAmountEth, balanceEth)
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		chainID, ChainIDErr = w.Client.NetworkID(ctx)
	}()

	go func() {
		defer wg.Done()
		nonce, NonceErr = w.Client.PendingNonceAt(ctx, w.From)
	}()

	wg.Wait()

	if ChainIDErr != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", ChainIDErr)
	}
	if NonceErr != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", NonceErr)
	}

	// Prepare transaction
	tx := types.NewTransaction(
		nonce,
		contract.Address,
		big.NewInt(0),
		gasLimit,
		gasPrice,
		txData,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), w.PrivKey)
	if err != nil {
		return nil, err
	}

	// Send
	err = w.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (w *WithdrawProcessor) TransferBNB(ctx context.Context, to common.Address, amount *decimal.Decimal) (*types.Transaction, error) {
	var wg sync.WaitGroup
	var balanceWithDecimals *big.Int
	var gasPrice *big.Int
	var gasLimit uint64
	var nonce uint64
	var chainID *big.Int

	var BalanceErr error
	var GasPriceErr error
	var GasLimitErr error
	var NonceErr error
	var ChainIDErr error

	wg.Add(3)
	amountWithDecimals := amount.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(18))).BigInt()
	go func() {
		defer wg.Done()
		balanceWithDecimals, BalanceErr = w.Client.BalanceAt(ctx, w.From, nil)
	}()

	go func() {
		defer wg.Done()
		gasPrice, GasPriceErr = w.Client.SuggestGasPrice(ctx)
	}()

	go func() {
		defer wg.Done()
		gasLimit, GasLimitErr = w.Client.EstimateGas(ctx, ethereum.CallMsg{
			From:  w.From,
			To:    &to,
			Value: amountWithDecimals,
			Data:  nil,
		})
	}()

	wg.Wait()

	if BalanceErr != nil {
		return nil, fmt.Errorf("failed to get balance: %v", BalanceErr)
	}
	if GasPriceErr != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", GasPriceErr)
	}
	if GasLimitErr != nil {
		re := regexp.MustCompile(`failed with (\d+) gas`)
		match := re.FindStringSubmatch(GasLimitErr.Error())
		if len(match) > 1 {
			gasLimit, _ = strconv.ParseUint(match[1], 10, 64)
		} else {
			return nil, fmt.Errorf("failed to estimate gas: %v", GasLimitErr)
		}
	}

	finalAmount := decimal.NewFromBigInt(new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice), int32(-18)).Add(*amount)
	log.Printf("finalAmount: %v, amount: %v, gasLimit: %v, gasPrice: %v", finalAmount, amount, gasLimit, gasPrice)
	balance := decimal.NewFromBigInt(balanceWithDecimals, int32(-18))
	log.Printf("balance: %v", balance)
	if balance.Cmp(finalAmount) < 0 {
		return nil, fmt.Errorf("insufficient balance. Need: %v, Have: %v", finalAmount, balance)
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		chainID, ChainIDErr = w.Client.NetworkID(ctx)
	}()

	go func() {
		defer wg.Done()
		nonce, NonceErr = w.Client.PendingNonceAt(ctx, w.From)
	}()

	wg.Wait()

	if ChainIDErr != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", ChainIDErr)
	}
	if NonceErr != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", NonceErr)
	}

	// Prepare transaction
	tx := types.NewTransaction(
		nonce,
		to,
		amountWithDecimals,
		gasLimit,
		gasPrice,
		nil,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), w.PrivKey)
	if err != nil {
		return nil, err
	}

	// Send
	err = w.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
