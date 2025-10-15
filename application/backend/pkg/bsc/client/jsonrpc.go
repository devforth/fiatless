package client

import (
	"context"
	"crypto/ecdsa"
	"fiatless/internal/models"
	"fiatless/pkg/bsc/address"
	"fiatless/pkg/httpclient"
	"fmt"
	"math/big"
	"net/http"

	fiatless_bsc "fiatless/pkg/bsc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
)

type BSCClient struct {
	ethClient     *ethclient.Client
	rateTransport *httpclient.RateLimitedTransport
}

// if rps > 0, the client will be rate limited
func NewBSCClient(url string, rps float64) (*BSCClient, error) {
	rateTransport := httpclient.NewRateLimitedTransport(rps, nil)
	rpcClient, err := rpc.DialOptions(context.Background(), url, rpc.WithHTTPClient(&http.Client{Transport: rateTransport}))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	return &BSCClient{
		ethClient:     ethClient,
		rateTransport: rateTransport,
	}, nil
}

func (c *BSCClient) GetBalance(address address.BSCAddress, includeBNB bool, bep20Tokens []address.BSCAddress) (models.BSCWalletBalance, error) {
	ctx := context.Background()
	result := models.BSCWalletBalance{}

	if includeBNB {
		ethAddr := common.BytesToAddress(address.Raw())
		balance, err := c.ethClient.BalanceAt(ctx, ethAddr, nil) // nil for latest block
		if err != nil {
			return models.BSCWalletBalance{}, fmt.Errorf("failed to get BNB balance: %v", err)
		}
		result.BNBBalance = decimal.NewFromBigInt(balance, -18)
	}

	for _, tokenAddr := range bep20Tokens {
		tokenBalance, err := c.getBEP20Balance(ctx, address, &tokenAddr)
		if err != nil {
			return models.BSCWalletBalance{}, fmt.Errorf("failed to get token balance for %s: %v", tokenAddr, err)
		}
		tokenBalanceDecimal, err := c.GetDecimals(ctx, &tokenAddr)
		if err != nil {
			return models.BSCWalletBalance{}, fmt.Errorf("failed to get decimals: %v", err)
		}
		balanceDecimal := decimal.NewFromBigInt(&tokenBalance, -int32(tokenBalanceDecimal))

		result.BEP20Balances = append(result.BEP20Balances, models.BEP20Balance{
			ContractAddress: tokenAddr,
			Balance:         balanceDecimal,
		})
	}

	return result, nil
}

func (c *BSCClient) getBEP20Balance(ctx context.Context, address address.BSCAddress, tokenAddress *address.BSCAddress) (big.Int, error) {
	contract, err := fiatless_bsc.NewBEP20Contract(c.ethClient, tokenAddress)
	if err != nil {
		return *big.NewInt(0), fmt.Errorf("failed to create contract: %v", err)
	}

	balanceOf, err := contract.Functions.Get("balanceOf")
	if err != nil {
		return *big.NewInt(0), fmt.Errorf("failed to get balanceOf function: %v", err)
	}

	walletAddress := common.HexToAddress(address.String())
	result, err := balanceOf.Call(ctx, walletAddress)
	if err != nil {
		return *big.NewInt(0), err
	}

	balance := result[0].(*big.Int)

	return *balance, nil
}

func (c *BSCClient) Withdraw(ctx context.Context, privateKey *ecdsa.PrivateKey, toAddress *address.BSCAddress, amount *decimal.Decimal, tokenAddress *address.BSCAddress) (models.BSCWithdrawResponse, error) {
	var tx *types.Transaction
	processor := fiatless_bsc.NewWithdrawProcessor(c.ethClient, privateKey)
	var transferErr error

	if tokenAddress != nil {
		contract, err := fiatless_bsc.NewBEP20Contract(c.ethClient, tokenAddress)
		if err != nil {
			return models.BSCWithdrawResponse{}, fmt.Errorf("failed to create contract: %v", err)
		}
		tx, transferErr = processor.TransferToken(ctx, contract, common.HexToAddress(toAddress.String()), amount)
	} else {
		tx, transferErr = processor.TransferBNB(ctx, common.HexToAddress(toAddress.String()), amount)
	}

	if transferErr != nil {
		return models.BSCWithdrawResponse{}, fmt.Errorf("failed to transfer: %v", transferErr)
	}

	return models.BSCWithdrawResponse{
		TransactionID: tx.Hash().Hex(),
	}, nil
}

func (c *BSCClient) GetDecimals(ctx context.Context, tokenAddress *address.BSCAddress) (uint8, error) {
	contract, err := fiatless_bsc.NewBEP20Contract(c.ethClient, tokenAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to create contract: %v", err)
	}

	decimals, err := contract.Functions.Get("decimals")
	if err != nil {
		return 0, fmt.Errorf("failed to get decimals: %v", err)
	}

	decimalsResult, err := decimals.Call(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get decimals: %v", err)
	}

	return decimalsResult[0].(uint8), nil
}

// Close closes the RPC client connection
func (c *BSCClient) Close() {
	if c.ethClient != nil {
		c.ethClient.Close()
	}
}

func (c *BSCClient) UpdateRateLimit(rps float64) {
	c.rateTransport.UpdateRPS(rps)
}
