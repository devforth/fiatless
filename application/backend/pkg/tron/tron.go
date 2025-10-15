package tron

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/client/grpc"
	"log"
	"math/big"
	"strconv"
	"strings"

	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

type TokenStruct struct {
	TokenType models.TokenType
	ID        uuid.UUID
}

type Tron struct {
	Client *grpc.TronClient
}

func NewTron(client *grpc.TronClient) *Tron {
	return &Tron{Client: client}
}

func (c *Tron) GetAccount(address *address.TronAddress) (*core_tronpb.Account, error) {
	decodedAddress := address.Bytes()

	account, err := c.Client.GetAccount(context.Background(), &core_tronpb.Account{Address: decodedAddress})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Tron) GetAccountResource(address *address.TronAddress) (*api_tronpb.AccountResourceMessage, error) {
	account, err := c.Client.GetAccountResource(context.Background(), &core_tronpb.Account{Address: address.Bytes()})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Tron) GetAccountTRXBalance(address *address.TronAddress) (decimal.Decimal, error) {
	account, err := c.GetAccount(address)
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	balance := decimal.NewFromBigInt(big.NewInt(account.Balance), -6)
	return balance, nil
}

func (c *Tron) GetAccountTRC20Balance(address *address.TronAddress, smartContractAddress *address.TronAddress) (decimal.Decimal, error) {
	ctx := context.Background()

	trc20, err := NewTRC20Contract(c.Client, smartContractAddress)
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	balanceFunc, err := trc20.Functions.Get("balanceOf")
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	balance, err := balanceFunc.Call(ctx, address.Bytes())
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	decimalsFunc, err := trc20.Functions.Get("decimals")
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	decimals, err := decimalsFunc.Call(ctx)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	scDecimal := decimal.NewFromInt(int64(decimals[0].(uint8)))

	balanceDecimal, err := decimal.NewFromString(balance[0].(string))
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	balanceDecimal = balanceDecimal.Div(decimal.NewFromInt(10).Pow(scDecimal))

	return balanceDecimal, nil
}

func (c *Tron) WithdrawTRX(privateKey *ecdsa.PrivateKey, toAddress *address.TronAddress, amount decimal.Decimal) (bp_models.TronWithdrawResponse, error) {
	ctx := context.Background()
	trxTx := NewWithdrawProcessor(c.Client, privateKey)
	return trxTx.Withdraw(ctx, toAddress, amount)
}

func (c *Tron) WithdrawTRC20(privateKey *ecdsa.PrivateKey, toAddress *address.TronAddress, amount decimal.Decimal, contractAddress *address.TronAddress) (bp_models.TronWithdrawResponse, error) {
	ctx := context.Background()
	trc20Tx := NewWithdrawProcessor(c.Client, privateKey)
	return trc20Tx.TransferTRC20Token(ctx, contractAddress, toAddress, amount)
}

func (c *Tron) GetLatestBlockNumber() (uint64, error) {
	if c.Client == nil {
		return 0, errors.New("client not initialized")
	}

	block, err := c.Client.GetNodeInfo(context.Background(), &api_tronpb.EmptyMessage{})
	if err != nil {
		return 0, err
	}
	prefix := "Num:"
	start := strings.Index(block.Block, prefix)
	if start == -1 {
		return 0, errors.New("num not found")
	}

	start += len(prefix)

	end := strings.IndexByte(block.Block[start:], ',')
	if end == -1 {
		return 0, errors.New("comma not found after Num")
	}

	num := block.Block[start : start+end]
	blockNumber, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func (c *Tron) GetBlocks(startBlockNumber uint64, endBlockNumber uint64) ([]*api_tronpb.BlockExtention, error) {
	blocks, err := c.Client.GetBlockByLimitNext2(context.Background(), &api_tronpb.BlockLimit{
		StartNum: int64(startBlockNumber),
		EndNum:   int64(endBlockNumber),
	})
	if err != nil {
		return nil, err
	}
	return blocks.Block, nil
}

func (c *Tron) ParseBlock(ctx context.Context, walletAddresses []address.TronAddress, tokens []bp_models.ParseBlocksToken, block *api_tronpb.BlockExtention) ([]bp_models.Deposit, error) {
	deposits := []bp_models.Deposit{}

	tokensMap := c.buildTokensMap(tokens)

	for _, transaction := range block.Transactions {
		if len(transaction.Transaction.RawData.Contract) == 0 {
			continue
		}

		contractType := transaction.Transaction.RawData.Contract[0].GetType()
		log.Println("contractType", contractType)
		switch contractType {
		case core_tronpb.Transaction_Contract_TransferContract:
			deposit, err := c.parseTransferContract(walletAddresses, tokensMap, transaction)
			if err != nil {
				log.Printf("Error parsing TransferContract: %v", err)
				continue
			}
			if deposit != nil {
				deposits = append(deposits, *deposit)
			}

		case core_tronpb.Transaction_Contract_TriggerSmartContract:
			deposit, err := c.parseTriggerSmartContract(ctx, walletAddresses, tokensMap, transaction)
			if err != nil {
				log.Printf("Error parsing TriggerSmartContract: %v", err)
				continue
			}
			if deposit != nil {
				deposits = append(deposits, *deposit)
			}

		default:
			log.Printf("Unknown contract type: %v", contractType)
			continue
		}
	}

	return deposits, nil
}

func (c *Tron) buildTokensMap(tokens []bp_models.ParseBlocksToken) map[string]TokenStruct {
	tokensMap := make(map[string]TokenStruct)

	for _, token := range tokens {
		switch token.TokenType {
		case models.TokenTypeNative:
			tokensMap["_trx_native"] = TokenStruct{
				TokenType: token.TokenType,
				ID:        token.ID,
			}
		case models.TokenTypeTRC10, models.TokenTypeTRC20:
			if token.TokenID != nil {
				tokensMap[*token.TokenID] = TokenStruct{
					TokenType: token.TokenType,
					ID:        token.ID,
				}
			}
		}
	}

	return tokensMap
}

func (c *Tron) parseTransferContract(walletAddresses []address.TronAddress, tokensMap map[string]TokenStruct, transaction *api_tronpb.TransactionExtention) (*bp_models.Deposit, error) {
	contractData := transaction.Transaction.RawData.Contract[0].Parameter.Value

	transferContract := &core_tronpb.TransferContract{}
	err := proto.Unmarshal(contractData, transferContract)
	if err != nil {
		return nil, err
	}

	toAddr, err := address.NewTronAddressFromBytes(transferContract.ToAddress)
	if err != nil {
		return nil, err
	}

	var matchedWallet *address.TronAddress
	for _, walletAddr := range walletAddresses {
		if walletAddr.String() == toAddr.String() {
			matchedWallet = &walletAddr
			break
		}
	}

	if matchedWallet == nil {
		return nil, nil
	}

	if _, exists := tokensMap["_trx_native"]; !exists {
		return nil, nil
	}

	deposit := &bp_models.Deposit{
		TxID:      hex.EncodeToString(transaction.Txid),
		TokenID:   tokensMap["_trx_native"].ID.String(),
		ToAddress: toAddr.String(),
		Amount:    decimal.NewFromBigInt(big.NewInt(transferContract.Amount), -6).String(),
		Timestamp: transaction.Transaction.RawData.Timestamp,
	}

	return deposit, nil
}

func (c *Tron) parseTriggerSmartContract(ctx context.Context, walletAddresses []address.TronAddress, tokensMap map[string]TokenStruct, transaction *api_tronpb.TransactionExtention) (*bp_models.Deposit, error) {
	contractData := transaction.Transaction.RawData.Contract[0].Parameter.Value
	log.Println("parseTriggerSmartContract")
	triggerContract := &core_tronpb.TriggerSmartContract{}
	err := proto.Unmarshal(contractData, triggerContract)
	if err != nil {
		return nil, err
	}
	log.Println("parseTriggerSmartContract 1")
	if len(triggerContract.Data) < 68 {
		return nil, nil
	}
	log.Println("parseTriggerSmartContract 2")
	methodSelector := hex.EncodeToString(triggerContract.Data[:4])
	if methodSelector != "a9059cbb" {
		return nil, nil
	}
	log.Println("parseTriggerSmartContract 3")
	toAddressBytes := triggerContract.Data[16:36]
	// Add TRON address prefix "41"
	toAddressBytesWithPrefix := make([]byte, 21)
	toAddressBytesWithPrefix[0] = 0x41
	copy(toAddressBytesWithPrefix[1:], toAddressBytes)
	log.Println("toAddressBytes", hex.EncodeToString(triggerContract.Data))
	toAddr, err := address.NewTronAddressFromBytes(toAddressBytesWithPrefix)
	if err != nil {
		return nil, err
	}
	log.Println("toAddr", toAddr.String())
	log.Println("parseTriggerSmartContract 4")
	var matchedWallet *address.TronAddress
	for _, walletAddr := range walletAddresses {
		if walletAddr.String() == toAddr.String() {
			matchedWallet = &walletAddr
			break
		}
	}

	if matchedWallet == nil {
		return nil, nil
	}
	log.Println("parseTriggerSmartContract 5")
	contractAddr, err := address.NewTronAddressFromBytes(triggerContract.ContractAddress)
	if err != nil {
		return nil, err
	}
	log.Println("parseTriggerSmartContract 6")
	token, exists := tokensMap[contractAddr.String()]
	if !exists || token.TokenType != models.TokenTypeTRC20 {
		return nil, nil
	}
	log.Println("parseTriggerSmartContract 7")
	amountBytes := triggerContract.Data[36:68]
	amount := new(big.Int).SetBytes(amountBytes)
	log.Println("parseTriggerSmartContract 8")

	log.Println("parseTriggerSmartContract 9")
	trc20Contract, err := NewTRC20Contract(c.Client, contractAddr)
	if err != nil {
		return nil, err
	}
	log.Println("parseTriggerSmartContract 10")
	decimalsFunc, err := trc20Contract.Functions.Get("decimals")
	if err != nil {
		return nil, err
	}
	log.Println("parseTriggerSmartContract 11")
	decimalsResult, err := decimalsFunc.Call(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("parseTriggerSmartContract 12")
	decimals := int32(decimalsResult[0].(uint8))
	log.Println("parseTriggerSmartContract 13")
	deposit := &bp_models.Deposit{
		TxID:      hex.EncodeToString(transaction.Txid),
		TokenID:   token.ID.String(),
		ToAddress: toAddr.String(),
		Amount:    decimal.NewFromBigInt(amount, -decimals).String(),
		Timestamp: transaction.Transaction.RawData.Timestamp,
	}
	log.Println("parseTriggerSmartContract 14")
	return deposit, nil
}
