package tron

import (
	"context"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/utils"
	"fmt"
	"strings"

	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var trc20ABI = `[
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
	Name     string
	Contract *Contract
	Client   api_tronpb.WalletClient
}

type Contract struct {
	Address   *address.TronAddress
	ABI       string
	Client    api_tronpb.WalletClient
	Functions *ContractFunctions
	parsedABI abi.ABI
}

func NewContract(client api_tronpb.WalletClient, address *address.TronAddress, abiJSON string) (*Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	c := &Contract{
		Address:   address,
		ABI:       abiJSON,
		Client:    client,
		parsedABI: parsedABI,
	}
	c.Functions = &ContractFunctions{contract: c}
	return c, nil
}

func (cf *ContractFunctions) Get(name string) (*ContractMethod, error) {
	_, exists := cf.contract.parsedABI.Methods[name]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found in ABI", name)
	}

	return &ContractMethod{
		Name:     name,
		Contract: cf.contract,
		Client:   cf.contract.Client,
	}, nil
}

// Call executes a read-only contract method
func (cm *ContractMethod) Call(ctx context.Context, args ...any) ([]any, error) {
	callParams, err := formatCallParams(cm.Name, cm.Contract.parsedABI, args...)
	if err != nil {
		return nil, err
	}

	req := &core_tronpb.TriggerSmartContract{
		OwnerAddress:    cm.Contract.Address.Bytes(),
		ContractAddress: cm.Contract.Address.Bytes(),
		Data:            callParams,
		CallValue:       0,
		TokenId:         0,
		CallTokenValue:  0,
	}

	result, err := cm.Client.TriggerConstantContract(ctx, req)
	if err != nil {
		return nil, err
	}

	return parseCallResult(result, cm.Contract.parsedABI, cm.Name)
}

func formatCallParams(methodName string, contractABI abi.ABI, args ...any) ([]byte, error) {
	method, exists := contractABI.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found in ABI", methodName)
	}

	data, err := method.Inputs.Pack(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments: %v", err)
	}

	selector := utils.GetFunctionSelector(method.Sig)

	return append(selector, data...), nil
}

func parseCallResult(result *api_tronpb.TransactionExtention, contractABI abi.ABI, methodName string) ([]any, error) {
	if result.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("call error: %s", string(result.GetResult().GetMessage()))
	}

	if len(result.GetConstantResult()) == 0 {
		return nil, fmt.Errorf("no constant result returned")
	}

	method, exists := contractABI.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found in ABI", methodName)
	}

	decoded, err := method.Outputs.Unpack(result.GetConstantResult()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return decoded, nil
}

func NewTRC20Contract(client api_tronpb.WalletClient, address *address.TronAddress) (*Contract, error) {
	return NewContract(client, address, trc20ABI)
}
