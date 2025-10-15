package utils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
)

func HexToBigInt(hexStr string) (*big.Int, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	n := new(big.Int)
	n, success := n.SetString(hexStr, 16)
	if !success {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return n, nil
}

func HexToDecimal(hexStr string) (*decimal.Decimal, error) {
	bigInt, err := HexToBigInt(hexStr)
	if err != nil {
		return nil, err
	}

	dec := decimal.NewFromBigInt(bigInt, 0)
	return &dec, nil
}
