package address

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
)

type BSCAddress struct {
	raw []byte
}

func NewBSCAddressFromHex(hexAddress string) (*BSCAddress, error) {
	hexAddress = strings.TrimPrefix(hexAddress, "0x")

	decoded, err := hex.DecodeString(hexAddress)
	if err != nil {
		return nil, err
	}

	if len(decoded) != 20 {
		return nil, errors.New("invalid address length, expected 20 bytes")
	}

	return &BSCAddress{
		raw: decoded,
	}, nil
}

func NewBSCAddressFromString(address string) (*BSCAddress, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}

	if !common.IsHexAddress(address) {
		return nil, errors.New("invalid bsc address format")
	}

	addr := common.HexToAddress(address)
	return &BSCAddress{
		raw: addr.Bytes(),
	}, nil
}

func NewBSCAddressFromBytes(address []byte) (*BSCAddress, error) {
	if len(address) != 20 {
		return nil, errors.New("invalid address length, expected 20 bytes")
	}

	return &BSCAddress{
		raw: address,
	}, nil
}

func NewBSCAddressFromPrivateKey(privateKey *ecdsa.PrivateKey) (*BSCAddress, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &BSCAddress{
		raw: address.Bytes(),
	}, nil
}

func NewBSCAddressFromBip32Key(bip32Key *bip32.Key) (*BSCAddress, error) {
	privateKey, err := crypto.ToECDSA(bip32Key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bip32 key to ECDSA: %v", err)
	}

	return NewBSCAddressFromPrivateKey(privateKey)
}

func (a *BSCAddress) Bytes() []byte {
	return a.raw
}

func (a *BSCAddress) Hex() string {
	return hex.EncodeToString(a.raw)
}

func (a *BSCAddress) String() string {
	return "0x" + hex.EncodeToString(a.raw)
}

func (a *BSCAddress) Raw() []byte {
	return a.raw
}

func (a *BSCAddress) ToChecksum() string {
	return common.HexToAddress(a.String()).Hex()
}

func (a *BSCAddress) Validate() error {
	if len(a.raw) != 20 {
		return fmt.Errorf("invalid address length: got %d, want 20", len(a.raw))
	}
	return nil
}

func (m BSCAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *BSCAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	log.Printf("UnmarshalJSON: %s", str)
	addr, err := NewBSCAddressFromString(str)
	if err != nil {
		return err
	}

	m.raw = addr.raw
	return nil
}
