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

type EthereumAddress struct {
	raw []byte
}

func NewEthereumAddressFromHex(hexAddress string) (*EthereumAddress, error) {
	hexAddress = strings.TrimPrefix(hexAddress, "0x")

	decoded, err := hex.DecodeString(hexAddress)
	if err != nil {
		return nil, err
	}

	if len(decoded) != 20 {
		return nil, errors.New("invalid address length, expected 20 bytes")
	}

	return &EthereumAddress{
		raw: decoded,
	}, nil
}

func NewEthereumAddressFromString(address string) (*EthereumAddress, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}

	if !common.IsHexAddress(address) {
		return nil, errors.New("invalid ethereum address format")
	}

	addr := common.HexToAddress(address)
	return &EthereumAddress{
		raw: addr.Bytes(),
	}, nil
}

func NewEthereumAddressFromBytes(address []byte) (*EthereumAddress, error) {
	if len(address) != 20 {
		return nil, errors.New("invalid address length, expected 20 bytes")
	}

	return &EthereumAddress{
		raw: address,
	}, nil
}

func NewEthereumAddressFromPrivateKey(privateKey *ecdsa.PrivateKey) (*EthereumAddress, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &EthereumAddress{
		raw: address.Bytes(),
	}, nil
}

func NewEthereumAddressFromBip32Key(bip32Key *bip32.Key) (*EthereumAddress, error) {
	privateKey, err := crypto.ToECDSA(bip32Key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bip32 key to ECDSA: %v", err)
	}

	return NewEthereumAddressFromPrivateKey(privateKey)
}

func (a *EthereumAddress) Bytes() []byte {
	return a.raw
}

func (a *EthereumAddress) Hex() string {
	return hex.EncodeToString(a.raw)
}

func (a *EthereumAddress) String() string {
	return "0x" + hex.EncodeToString(a.raw)
}

func (a *EthereumAddress) Raw() []byte {
	return a.raw
}

func (a *EthereumAddress) ToChecksum() string {
	return common.HexToAddress(a.String()).Hex()
}

func (a *EthereumAddress) Validate() error {
	if len(a.raw) != 20 {
		return fmt.Errorf("invalid address length: got %d, want 20", len(a.raw))
	}
	return nil
}

func (m EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *EthereumAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	log.Printf("UnmarshalJSON: %s", str)
	addr, err := NewEthereumAddressFromString(str)
	if err != nil {
		return err
	}

	m.raw = addr.raw
	return nil
}
