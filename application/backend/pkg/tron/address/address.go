package address

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fiatless/pkg/utils"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/crypto/sha3"
)

type TronAddress struct {
	raw []byte
}

func NewTronAddressFromBase58(address string) (*TronAddress, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}

	log.Printf("Parsing TRON address: %s", address)

	decoded, _, err := base58.CheckDecode(address)
	if err != nil {
		return nil, fmt.Errorf("base58check decode error: %v", err)
	}

	if len(decoded) == 20 {
		decoded = append([]byte{0x41}, decoded...)
	}

	if len(decoded) != 21 {
		return nil, fmt.Errorf("invalid address length: got %d, want 21", len(decoded))
	}

	if decoded[0] != 0x41 {
		return nil, fmt.Errorf("invalid address prefix: got 0x%x, want 0x41", decoded[0])
	}

	addr := &TronAddress{
		raw: decoded,
	}

	if err := addr.Validate(); err != nil {
		return nil, err
	}

	return addr, nil
}

func NewTronAddressFromHex(hexAddress string) (*TronAddress, error) {
	hexAddress = strings.TrimPrefix(hexAddress, "0x")

	decoded, err := hex.DecodeString(hexAddress)
	if err != nil {
		return nil, err
	}

	if len(decoded) != 21 {
		return nil, errors.New("invalid address length, expected 21 bytes")
	}

	return &TronAddress{
		raw: decoded,
	}, nil
}

func NewTronAddressFromBytes(address []byte) (*TronAddress, error) {
	return &TronAddress{
		raw: address,
	}, nil
}

func NewTronAddressFromPrivateKey(privateKey *ecdsa.PrivateKey) (*TronAddress, error) {
	pubKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)

	pubKeyBytes = pubKeyBytes[1:]

	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubKeyBytes)
	pubHash := hash.Sum(nil)

	tronAddress := append([]byte{0x41}, pubHash[len(pubHash)-20:]...)
	return &TronAddress{
		raw: tronAddress,
	}, nil
}

func NewTronAddressFromBip32Key(bip32Key *bip32.Key) (*TronAddress, error) {
	ecdsaKey, err := utils.NewECDSAKeyFromBip32Key(bip32Key)
	if err != nil {
		return nil, err
	}

	return NewTronAddressFromPrivateKey(ecdsaKey)
}

func (a *TronAddress) Bytes() []byte {
	return a.raw
}

func (a *TronAddress) Base58() string {
	return base58.Encode(a.raw)
}

func (a *TronAddress) Hex() string {
	return hex.EncodeToString(a.raw)
}

func (a *TronAddress) Equals(other *TronAddress) bool {
	return bytes.Equal(a.raw, other.raw)
}

func (a *TronAddress) String() string {
	firstHash := sha256.Sum256(a.raw)
	secondHash := sha256.Sum256(firstHash[:])

	rawWithChecksum := append(a.raw, secondHash[:4]...)

	return base58.Encode(rawWithChecksum)
}

func (a *TronAddress) Raw() []byte {
	return a.raw
}

func (a *TronAddress) Validate() error {
	if len(a.raw) != 21 {
		return fmt.Errorf("invalid address length: got %d, want 21", len(a.raw))
	}
	if a.raw[0] != 0x41 {
		return fmt.Errorf("invalid address prefix: got 0x%x, want 0x41", a.raw[0])
	}
	return nil
}

func (a TronAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *TronAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	addr, err := NewTronAddressFromBase58(str)
	if err != nil {
		return err
	}

	a.raw = addr.raw
	return nil
}
