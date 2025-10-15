package utils

import (
	"crypto/ecdsa"
	api_tronpb "fiatless/pkg/proto/tron/api"
	core_tronpb "fiatless/pkg/proto/tron/core"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/wallet"
	"fmt"
	"math/big"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type TronTransactionBuilder struct {
	privateKey  *ecdsa.PrivateKey
	fromAddress *address.TronAddress
	toAddress   *address.TronAddress
	amount      *big.Int

	refBlockBytes  []byte
	refBlockNum    int64
	refBlockHash   []byte
	expirationTime int64
	feeLimit       int64

	contractType core_tronpb.Transaction_Contract_ContractType
	contractData proto.Message

	transaction *core_tronpb.Transaction
}

func NewTronTransactionBuilder(privateKey *ecdsa.PrivateKey) *TronTransactionBuilder {
	fromWallet := wallet.NewBaseTronWallet(privateKey)
	fromAddress := fromWallet.GetAddress()

	return &TronTransactionBuilder{
		privateKey:     privateKey,
		fromAddress:    fromAddress,
		feeLimit:       0,
		expirationTime: time.Now().UnixMilli() + 60*60*1000,
		contractType:   core_tronpb.Transaction_Contract_TransferContract,
	}
}

func (b *TronTransactionBuilder) SetReferenceBlock(blockExtension *api_tronpb.BlockExtention) *TronTransactionBuilder {
	blockID := blockExtension.Blockid

	b.refBlockBytes = blockID[6:8]
	b.refBlockNum = blockExtension.BlockHeader.RawData.Number
	b.refBlockHash = blockID[8:16]

	return b
}

func (b *TronTransactionBuilder) SetReferenceBlockManual(refBlockBytes []byte, refBlockNum int64, refBlockHash []byte) *TronTransactionBuilder {
	b.refBlockBytes = refBlockBytes
	b.refBlockNum = refBlockNum
	b.refBlockHash = refBlockHash

	return b
}

func (b *TronTransactionBuilder) SetContractType(contractType core_tronpb.Transaction_Contract_ContractType) *TronTransactionBuilder {
	b.contractType = contractType

	return b
}

func (b *TronTransactionBuilder) SetContractData(contractData proto.Message) *TronTransactionBuilder {
	b.contractData = contractData

	return b
}

func (b *TronTransactionBuilder) SetExpiration(expirationTime int64) *TronTransactionBuilder {
	b.expirationTime = expirationTime

	return b
}

func (b *TronTransactionBuilder) SetFeeLimit(feeLimit int64) *TronTransactionBuilder {
	b.feeLimit = feeLimit

	return b
}

func (b *TronTransactionBuilder) TransferTRX(toAddress *address.TronAddress, amount *big.Int) *TronTransactionBuilder {
	b.toAddress = toAddress
	b.amount = amount
	b.contractType = core_tronpb.Transaction_Contract_TransferContract

	b.contractData = &core_tronpb.TransferContract{
		OwnerAddress: b.fromAddress.Bytes(),
		ToAddress:    toAddress.Bytes(),
		Amount:       amount.Int64(),
	}

	return b
}

func (b *TronTransactionBuilder) Build() (*core_tronpb.Transaction, error) {
	if b.refBlockBytes == nil || b.refBlockHash == nil {
		return nil, ErrMissingBlockReference
	}

	var contractParameter *anypb.Any
	var err error

	switch b.contractType {
	case core_tronpb.Transaction_Contract_TransferContract:
		contractParameter, err = anypb.New(b.contractData.(*core_tronpb.TransferContract))
	case core_tronpb.Transaction_Contract_TriggerSmartContract:
		contractParameter, err = anypb.New(b.contractData.(*core_tronpb.TriggerSmartContract))
	default:
		return nil, ErrUnsupportedContractType
	}

	if err != nil {
		return nil, err
	}

	rawData := &core_tronpb.TransactionRaw{
		RefBlockBytes: b.refBlockBytes,
		RefBlockHash:  b.refBlockHash,
		Expiration:    b.expirationTime,
		FeeLimit:      b.feeLimit,
		Timestamp:     time.Now().UnixMilli(),
		Contract: []*core_tronpb.Transaction_Contract{
			{
				Type:      b.contractType,
				Parameter: contractParameter,
			},
		},
	}

	b.transaction = &core_tronpb.Transaction{
		RawData: rawData,
	}

	return b.transaction, nil
}

func (b *TronTransactionBuilder) Sign() (*core_tronpb.Transaction, error) {
	if b.transaction == nil {
		_, err := b.Build()
		if err != nil {
			return nil, err
		}
	}

	rawDataBytes, err := proto.Marshal(b.transaction.RawData)
	if err != nil {
		return nil, err
	}

	fromWallet := wallet.NewBaseTronWallet(b.privateKey)
	signature, err := fromWallet.SignTransaction(rawDataBytes)
	if err != nil {
		return nil, err
	}

	b.transaction.Signature = append(b.transaction.Signature, signature)

	return b.transaction, nil
}

func (b *TronTransactionBuilder) GetTransactionID() (string, error) {
	if b.transaction == nil {
		return "", ErrTransactionNotBuilt
	}

	return GetTxID(b.transaction)
}

func (b *TronTransactionBuilder) ToBytes() ([]byte, error) {
	if b.transaction == nil {
		return nil, ErrTransactionNotBuilt
	}

	return proto.Marshal(b.transaction)
}

func (b *TronTransactionBuilder) CalculateBandwidth() (int64, error) {
	if b.transaction == nil {
		return 0, ErrTransactionNotBuilt
	}

	if len(b.transaction.Signature) == 0 {
		return 0, fmt.Errorf("transaction is not signed yet")
	}

	rawDataBytes, err := proto.Marshal(b.transaction)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal raw data: %v", err)
	}

	bandwidth := EstimateBandwidth(rawDataBytes)

	return bandwidth, nil
}

func (b *TronTransactionBuilder) FromBytes(data []byte) error {
	tx := &core_tronpb.Transaction{}
	if err := proto.Unmarshal(data, tx); err != nil {
		return err
	}

	b.transaction = tx
	return nil
}

var (
	ErrMissingBlockReference   = fmt.Errorf("missing block reference data")
	ErrUnsupportedContractType = fmt.Errorf("unsupported contract type")
	ErrTransactionNotBuilt     = fmt.Errorf("transaction not built yet")
)

func EstimateBandwidth(rawData []byte) int64 {
	const (
		MAX_RESULT_SIZE_IN_TX = 64 // Maximum result size in transaction
	)

	return int64(len(rawData) + MAX_RESULT_SIZE_IN_TX)
}
