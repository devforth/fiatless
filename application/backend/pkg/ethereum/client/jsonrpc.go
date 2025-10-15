package client

import (
	"context"
	"crypto/ecdsa"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/ethereum/address"
	"fiatless/pkg/httpclient"
	"fmt"
	"math/big"
	"net/http"

	fiatless_ethereum "fiatless/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
)

type EthereumClient struct {
	rpcClient     *rpc.Client
	ethClient     *ethclient.Client
	rateTransport *httpclient.RateLimitedTransport
}

// if rps > 0, the client will be rate limited
func NewEthereumClient(url string, rps float64) (*EthereumClient, error) {
	rateTransport := httpclient.NewRateLimitedTransport(rps, nil)
	rpcClient, err := rpc.DialOptions(context.Background(), url, rpc.WithHTTPClient(&http.Client{Transport: rateTransport}))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}
	// Create ethclient
	ethClient := ethclient.NewClient(rpcClient)

	return &EthereumClient{
		rpcClient:     rpcClient,
		ethClient:     ethClient,
		rateTransport: rateTransport,
	}, nil
}

func (c *EthereumClient) GetBalance(address address.EthereumAddress, includeETH bool, erc20Tokens []address.EthereumAddress) (models.EthereumWalletBalance, error) {
	ctx := context.Background()
	result := models.EthereumWalletBalance{}

	if includeETH {
		ethAddr := common.BytesToAddress(address.Raw())
		balance, err := c.ethClient.BalanceAt(ctx, ethAddr, nil) // nil for latest block
		if err != nil {
			return models.EthereumWalletBalance{}, fmt.Errorf("failed to get ETH balance: %v", err)
		}
		result.ETHBalance = decimal.NewFromBigInt(balance, -18)
	}

	for _, tokenAddr := range erc20Tokens {
		tokenBalance, err := c.getERC20Balance(ctx, address, &tokenAddr)
		if err != nil {
			return models.EthereumWalletBalance{}, fmt.Errorf("failed to get token balance for %s: %v", tokenAddr, err)
		}
		tokenBalanceDecimal, err := c.GetDecimals(ctx, &tokenAddr)
		if err != nil {
			return models.EthereumWalletBalance{}, fmt.Errorf("failed to get decimals: %v", err)
		}
		balanceDecimal := decimal.NewFromBigInt(&tokenBalance, -int32(tokenBalanceDecimal))

		result.ERC20Balances = append(result.ERC20Balances, models.ERC20Balance{
			ContractAddress: tokenAddr,
			Balance:         balanceDecimal,
		})
	}

	return result, nil
}

// getERC20Balance gets the balance of an ERC20 token for a specific address
func (c *EthereumClient) getERC20Balance(ctx context.Context, address address.EthereumAddress, tokenAddress *address.EthereumAddress) (big.Int, error) {
	contract, err := fiatless_ethereum.NewERC20Contract(c.ethClient, tokenAddress)
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

func (c *EthereumClient) Withdraw(ctx context.Context, privateKey *ecdsa.PrivateKey, toAddress *address.EthereumAddress, amount *decimal.Decimal, tokenAddress *address.EthereumAddress) (bp_models.EthereumWithdrawResponse, error) {
	var tx *types.Transaction
	processor := fiatless_ethereum.NewWithdrawProcessor(c.ethClient, privateKey)
	var transferErr error

	if tokenAddress != nil {
		contract, err := fiatless_ethereum.NewERC20Contract(c.ethClient, tokenAddress)
		if err != nil {
			return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("failed to create contract: %v", err)
		}
		tx, transferErr = processor.TransferToken(ctx, contract, common.HexToAddress(toAddress.String()), amount)
	} else {
		tx, transferErr = processor.TransferETH(ctx, common.HexToAddress(toAddress.String()), amount)
	}

	if transferErr != nil {
		return bp_models.EthereumWithdrawResponse{}, fmt.Errorf("failed to transfer: %v", transferErr)
	}

	return bp_models.EthereumWithdrawResponse{
		TransactionID: tx.Hash().Hex(),
	}, nil
}

func (c *EthereumClient) GetDecimals(ctx context.Context, tokenAddress *address.EthereumAddress) (uint8, error) {
	contract, err := fiatless_ethereum.NewERC20Contract(c.ethClient, tokenAddress)
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
func (c *EthereumClient) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}

func (c *EthereumClient) UpdateRateLimit(rps float64) {
	c.rateTransport.UpdateRPS(rps)
}
