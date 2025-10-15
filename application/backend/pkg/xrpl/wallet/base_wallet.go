package wallet

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Peersyst/xrpl-go/keypairs"
	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	xrpl_wallet "github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"github.com/tyler-smith/go-bip32"
)

type BaseXRPLWallet struct {
	wallet *xrpl_wallet.Wallet
}

func (w *BaseXRPLWallet) GetAddress() xrpl_types.Address {
	return w.wallet.ClassicAddress
}
func (w *BaseXRPLWallet) SignTransaction(tx map[string]interface{}) (string, error) {
	txBlob, _, err := w.wallet.Sign(tx)
	if err != nil {
		return "", err
	}
	return txBlob, nil
}

// FromBip32KeyString creates an XRPL wallet from a private key string (adapted to work with GetPrivateKey format)
func FromBip32KeyString(privateKeyStr string) *xrpl_wallet.Wallet {
	// Use the new FromPrivateKey method which handles the format correctly
	wallet, err := FromPrivateKey(privateKeyStr)
	if err != nil {
		// Fallback to the original logic if the new method fails
		// This maintains backward compatibility
		privKey := strings.ToUpper(base58.Encode([]byte(privateKeyStr)))
		pubKey := strings.ToUpper(base58.Encode([]byte(privateKeyStr)))
		classicAddr, _ := keypairs.DeriveClassicAddress(pubKey)
		return &xrpl_wallet.Wallet{
			PrivateKey:     fmt.Sprintf("00%s", privKey),
			PublicKey:      pubKey,
			Seed:           "",
			ClassicAddress: xrpl_types.Address(classicAddr),
		}
	}
	return wallet
}

// FromPrivateKey creates an XRPL wallet from a private key string (as returned by GetPrivateKey)
func FromPrivateKey(privateKeyStr string) (*xrpl_wallet.Wallet, error) {
	// Remove the "00" prefix if present
	privateKeyStr = strings.TrimPrefix(privateKeyStr, "00")

	// Decode the hex private key
	privKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %v", err)
	}

	// Convert to ECDSA private key to derive the public key
	ecdsaPrivKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ECDSA private key: %v", err)
	}

	// Derive the compressed public key (XRPL requires compressed format)
	pubKeyBytes := crypto.CompressPubkey(&ecdsaPrivKey.PublicKey)

	privKey := strings.ToUpper(hex.EncodeToString(privKeyBytes))
	pubKey := strings.ToUpper(hex.EncodeToString(pubKeyBytes))
	classicAddr, err := keypairs.DeriveClassicAddress(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive classic address: %v", err)
	}

	return &xrpl_wallet.Wallet{
		PrivateKey:     fmt.Sprintf("00%s", privKey),
		PublicKey:      pubKey,
		Seed:           "",
		ClassicAddress: xrpl_types.Address(classicAddr),
	}, nil
}

func FromBip32Key(bip32Key *bip32.Key) *xrpl_wallet.Wallet {
	privKey := strings.ToUpper(hex.EncodeToString(bip32Key.Key))
	pubKey := strings.ToUpper(hex.EncodeToString(bip32Key.PublicKey().Key))
	classicAddr, _ := keypairs.DeriveClassicAddress(pubKey)
	return &xrpl_wallet.Wallet{
		PrivateKey:     fmt.Sprintf("00%s", privKey),
		PublicKey:      pubKey,
		Seed:           "",
		ClassicAddress: xrpl_types.Address(classicAddr),
	}
}

func NewBaseXRPLWallet(bip32Key *bip32.Key) *BaseXRPLWallet {
	return &BaseXRPLWallet{
		wallet: FromBip32Key(bip32Key),
	}
}

// NewBaseXRPLWalletFromPrivateKey creates a BaseXRPLWallet from a private key string
func NewBaseXRPLWalletFromPrivateKey(privateKeyStr string) (*BaseXRPLWallet, error) {
	wallet, err := FromPrivateKey(privateKeyStr)
	if err != nil {
		return nil, err
	}
	return &BaseXRPLWallet{
		wallet: wallet,
	}, nil
}

func (w *BaseXRPLWallet) GetPrivateKey() string {
	return w.wallet.PrivateKey
}
