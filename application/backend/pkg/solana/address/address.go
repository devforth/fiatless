package address

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
)

// SolanaAddress represents a Solana address which is the 32-byte Ed25519 public key.
type SolanaAddress struct {
	raw []byte
}

// NewSolanaAddressFromBase58 decodes a base58-encoded Solana address to raw 32 bytes.
func NewSolanaAddressFromBase58(addr string) (*SolanaAddress, error) {
	if addr == "" {
		return nil, errors.New("empty address")
	}

	decoded, err := base58.Decode(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid base58 address: %w", err)
	}
	if len(decoded) != 32 {
		return nil, fmt.Errorf("invalid address length: got %d, want 32", len(decoded))
	}

	return &SolanaAddress{raw: decoded}, nil
}

// NewSolanaAddressFromString is an alias of NewSolanaAddressFromBase58.
func NewSolanaAddressFromString(address string) (*SolanaAddress, error) {
	return NewSolanaAddressFromBase58(address)
}

// NewSolanaAddressFromHex accepts a hex string of 32 bytes (non-standard for Solana).
func NewSolanaAddressFromHex(hexAddress string) (*SolanaAddress, error) {
	decoded, err := hex.DecodeString(hexAddress)
	if err != nil {
		return nil, err
	}
	if len(decoded) != 32 {
		return nil, errors.New("invalid address length, expected 32 bytes")
	}
	return &SolanaAddress{raw: decoded}, nil
}

// NewSolanaAddressFromBytes constructs from raw 32 bytes.
func NewSolanaAddressFromBytes(address []byte) (*SolanaAddress, error) {
	if len(address) != 32 {
		return nil, errors.New("invalid address length, expected 32 bytes")
	}
	return &SolanaAddress{raw: address}, nil
}

// NewSolanaAddressFromPrivateKey derives the public address from an Ed25519 private key.
func NewSolanaAddressFromPrivateKey(privateKey ed25519.PrivateKey) (*SolanaAddress, error) {
	// ed25519.PrivateKey.Public() returns ed25519.PublicKey (32 bytes)
	pub, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("failed to get ed25519 public key")
	}
	if len(pub) != 32 {
		return nil, fmt.Errorf("unexpected public key length: %d", len(pub))
	}
	return &SolanaAddress{raw: pub}, nil
}

func (a *SolanaAddress) Bytes() []byte {
	return a.raw
}

func (a *SolanaAddress) Hex() string {
	return hex.EncodeToString(a.raw)
}

func (a *SolanaAddress) String() string {
	return base58.Encode(a.raw)
}

func (a *SolanaAddress) Raw() []byte {
	return a.raw
}

// ToChecksum returns the canonical string representation. Solana addresses have no EIP-55 checksum.
func (a *SolanaAddress) ToChecksum() string {
	return a.String()
}

func (a *SolanaAddress) Validate() error {
	if len(a.raw) != 32 {
		return fmt.Errorf("invalid address length: got %d, want 32", len(a.raw))
	}
	return nil
}

func (m SolanaAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *SolanaAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	addr, err := NewSolanaAddressFromString(str)
	if err != nil {
		return err
	}
	m.raw = addr.raw
	return nil
}
