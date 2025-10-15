package bsc

import (
	"context"
	"fiatless/pkg/bsc/address"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var bep20ABI = `[
  {"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},
  {"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"type":"function"},
  {"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"},
  {"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"type":"function"},
  {"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},
  {"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},
  {"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},
  {"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"},
  {"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},
  {"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},
  {"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}
]`

type ContractFunctions struct {
	contract *Contract
}

type ContractMethod struct {
	ABIEntry abi.Method
	Contract *Contract
}
type Contract struct {
	Address   common.Address
	ABI       abi.ABI
	Client    *ethclient.Client
	Functions *ContractFunctions
}

func NewContract(client *ethclient.Client, address string, abiJSON string) (*Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	c := &Contract{
		Address: common.HexToAddress(address),
		ABI:     parsedABI,
		Client:  client,
	}
	c.Functions = &ContractFunctions{contract: c}
	return c, nil
}

func (cf *ContractFunctions) Get(name string) (*ContractMethod, error) {
	method, exists := cf.contract.ABI.Methods[name]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found in ABI", name)
	}
	return &ContractMethod{
		ABIEntry: method,
		Contract: cf.contract,
	}, nil
}

// --- Call method (read-only/view) ---
func (cm *ContractMethod) Call(ctx context.Context, args ...any) ([]any, error) {
	data, err := cm.ABIEntry.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}

	fullData := append(cm.ABIEntry.ID, data...)

	callMsg := ethereum.CallMsg{
		To:   &cm.Contract.Address,
		Data: fullData,
	}

	// Static eth_call
	output, err := cm.Contract.Client.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, err
	}

	return cm.ABIEntry.Outputs.UnpackValues(output)
}

func NewBEP20Contract(client *ethclient.Client, address *address.BSCAddress) (*Contract, error) {
	return NewContract(client, address.String(), bep20ABI)
}
