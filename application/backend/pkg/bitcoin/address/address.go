package address

import (
	"encoding/json"
	"errors"
	"fiatless/internal/models"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/tyler-smith/go-bip32"
)

// BitcoinAddressType represents the type of Bitcoin address
type BitcoinAddressType int

const (
	P2PKH  BitcoinAddressType = iota // Pay-to-Public-Key-Hash (Legacy)
	P2SH                             // Pay-to-Script-Hash (SegWit wrapped)
	P2WPKH                           // Pay-to-Witness-Public-Key-Hash (Native SegWit)
	P2TR                             // Pay-to-Taproot
)

type BitcoinAddress struct {
	address     btcutil.Address
	addressType BitcoinAddressType
	network     models.BitcoinNetwork
}

// NewBitcoinAddressFromString creates a Bitcoin address from a string representation
func NewBitcoinAddressFromString(addressStr string, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	if addressStr == "" {
		return nil, errors.New("empty address")
	}

	btcutilNetwork := toBtcutilNetwork(network)

	addr, err := btcutil.DecodeAddress(addressStr, btcutilNetwork)
	if err != nil {
		return nil, fmt.Errorf("invalid bitcoin address: %v", err)
	}

	addressType := getAddressType(addr)

	return &BitcoinAddress{
		address:     addr,
		addressType: addressType,
		network:     network,
	}, nil
}

func toBtcutilNetwork(network models.BitcoinNetwork) *chaincfg.Params {
	switch network {
	case models.BitcoinMainnet:
		return &chaincfg.MainNetParams
	case models.BitcoinSignet:
		return &chaincfg.SigNetParams
	default:
		return &chaincfg.MainNetParams
	}
}

// NewBitcoinAddressFromPrivateKey creates a Bitcoin address from an ECDSA private key
func NewBitcoinAddressFromPrivateKey(privateKey *btcec.PrivateKey, addressType BitcoinAddressType, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	btcutilNetwork := toBtcutilNetwork(network)

	publicKey := privateKey.PubKey()
	pubKeyBytes := publicKey.SerializeCompressed()

	// Create the appropriate address type
	var addr btcutil.Address
	var err error

	switch addressType {
	case P2PKH:
		pubKeyHash := btcutil.Hash160(pubKeyBytes)
		addr, err = btcutil.NewAddressPubKeyHash(pubKeyHash, btcutilNetwork)
	case P2SH:
		pubKeyHash := btcutil.Hash160(pubKeyBytes)
		witnessProgram := append([]byte{0x00, 0x14}, pubKeyHash...)
		scriptHash := btcutil.Hash160(witnessProgram)
		addr, err = btcutil.NewAddressScriptHash(scriptHash, btcutilNetwork)
	case P2WPKH:
		pubKeyHash := btcutil.Hash160(pubKeyBytes)
		addr, err = btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, btcutilNetwork)
	case P2TR:
		// BIP86: compute tweaked output x-only from the private key
		tweakedXOnly, errTweaked := computeBIP86TweakedXOnly(privateKey)
		if errTweaked != nil {
			return nil, fmt.Errorf("failed to compute BIP86 tweak: %v", errTweaked)
		}
		addr, err = btcutil.NewAddressTaproot(tweakedXOnly, btcutilNetwork)
	default:
		return nil, errors.New("unsupported address type")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create address: %v", err)
	}

	return &BitcoinAddress{
		address:     addr,
		addressType: addressType,
		network:     network,
	}, nil
}

// NewBitcoinAddressFromBip32Key creates a Bitcoin address from a BIP32 key
func NewBitcoinAddressFromBip32Key(bip32Key *bip32.Key, addressType BitcoinAddressType, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	privateKey, err := btcec.PrivKeyFromBytes(bip32Key.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bip32 key to ECDSA: %v", err)
	}

	return NewBitcoinAddressFromPrivateKey(privateKey, addressType, network)
}

func computeBIP86TweakedXOnly(priv *btcec.PrivateKey) ([]byte, error) {
	outKey := txscript.ComputeTaprootKeyNoScript(priv.PubKey())
	return schnorr.SerializePubKey(outKey), nil
}

// String returns the string representation of the Bitcoin address
func (a *BitcoinAddress) String() string {
	return a.address.String()
}

// Bytes returns the raw bytes of the address
func (a *BitcoinAddress) Bytes() []byte {
	return a.address.ScriptAddress()
}

// GetAddressType returns the type of the Bitcoin address
func (a *BitcoinAddress) GetAddressType() BitcoinAddressType {
	return a.addressType
}

// GetNetwork returns the network parameters for this address
func (a *BitcoinAddress) GetNetwork() models.BitcoinNetwork {
	return a.network
}

// IsMainNet returns true if the address is for mainnet
func (a *BitcoinAddress) IsMainNet() bool {
	return a.network == models.BitcoinMainnet
}

// IsTestNet returns true if the address is for testnet
func (a *BitcoinAddress) IsTestNet() bool {
	return a.network == models.BitcoinSignet
}

// IsRegTest returns true if the address is for regtest
func (a *BitcoinAddress) IsRegTest() bool {
	return a.network == models.BitcoinSignet
}

// GetScriptPubKey returns the script public key for this address
func (a *BitcoinAddress) GetScriptPubKey() ([]byte, error) {
	return txscript.PayToAddrScript(a.address)
}

// Validate validates the Bitcoin address
func (a *BitcoinAddress) Validate() error {
	if a.address == nil {
		return errors.New("address is nil")
	}

	// Check if address is valid for the network
	if !a.address.IsForNet(toBtcutilNetwork(a.network)) {
		return fmt.Errorf("address is not valid for network %s", a.network)
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (a BitcoinAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (a *BitcoinAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	log.Printf("UnmarshalJSON: %s", str)

	addr, err := NewBitcoinAddressFromString(str, models.BitcoinSignet)
	if err != nil {
		return err
	}

	a.address = addr.address
	a.addressType = addr.addressType
	a.network = addr.network
	return nil
}

// getAddressType determines the address type from a btcutil.Address
func getAddressType(addr btcutil.Address) BitcoinAddressType {
	switch addr.(type) {
	case *btcutil.AddressTaproot:
		return P2TR
	case *btcutil.AddressPubKeyHash:
		return P2PKH
	case *btcutil.AddressScriptHash:
		return P2SH
	case *btcutil.AddressWitnessPubKeyHash:
		return P2WPKH
	default:
		return P2PKH // Default to P2PKH for unknown types
	}
}

// Helper functions for creating addresses with specific networks

// NewBitcoinAddressMainNet creates a Bitcoin address for mainnet
func NewBitcoinAddressMainNet(addressStr string) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromString(addressStr, models.BitcoinMainnet)
}

// NewBitcoinAddressTestNet creates a Bitcoin address for testnet
func NewBitcoinAddressTestNet(addressStr string) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromString(addressStr, models.BitcoinSignet)
}

// NewBitcoinAddressRegTest creates a Bitcoin address for regtest
func NewBitcoinAddressRegTest(addressStr string) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromString(addressStr, models.BitcoinSignet)
}

// NewP2PKHAddress creates a new P2PKH address from private key
func NewP2PKHAddress(privateKey *btcec.PrivateKey, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromPrivateKey(privateKey, P2PKH, network)
}

// NewP2SHAddress creates a new P2SH address from private key
func NewP2SHAddress(privateKey *btcec.PrivateKey, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromPrivateKey(privateKey, P2SH, network)
}

// NewP2WPKHAddress creates a new P2WPKH address from private key
func NewP2WPKHAddress(privateKey *btcec.PrivateKey, network models.BitcoinNetwork) (*BitcoinAddress, error) {
	return NewBitcoinAddressFromPrivateKey(privateKey, P2WPKH, network)
}
