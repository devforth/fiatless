package config

import (
	"context"
	"fiatless/internal/models"
	"fiatless/pkg/bitcoin/address"
	"fmt"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

type Mode string

const (
	ModeTestnet Mode = "testnet"
	ModeMainnet Mode = "mainnet"
)

type EnvConfig struct {
	Port               string `env:"PORT, default=8080"`
	IJSONEndpoint      string `env:"IJSON_ENDPOINT, default=http://localhost:8001"`
	DatabaseDSN        string `env:"DATABASE_DSN, default=host=localhost user=fiatless password=fiatless dbname=fiatless port=5432 sslmode=disable TimeZone=UTC"`
	TronMnemonic       string `env:"TRON_WALLET_MNEMONIC"`
	TronPassphrase     string `env:"TRON_WALLET_PASSPHRASE"`
	EthereumMnemonic   string `env:"ETHEREUM_WALLET_MNEMONIC"`
	EthereumPassphrase string `env:"ETHEREUM_WALLET_PASSPHRASE"`
	BSCMnemonic        string `env:"BSC_WALLET_MNEMONIC"`
	BSCPassphrase      string `env:"BSC_WALLET_PASSPHRASE"`
	BitcoinMnemonic    string `env:"BITCOIN_WALLET_MNEMONIC"`
	BitcoinPassphrase  string `env:"BITCOIN_WALLET_PASSPHRASE"`
	BitcoinAddressType string `env:"BITCOIN_WALLET_ADDRESS_TYPE, default=p2tr"`
	SolanaMnemonic     string `env:"SOLANA_WALLET_MNEMONIC"`
	SolanaPassphrase   string `env:"SOLANA_WALLET_PASSPHRASE"`
	XRPLMnemonic       string `env:"XRPL_WALLET_MNEMONIC"`
	XRPLPassphrase     string `env:"XRPL_WALLET_PASSPHRASE"`
	Mode               Mode   `env:"MODE, default=testnet"`
}

type Config struct {
	Port               string
	IJSONEndpoint      string
	DatabaseDSN        string
	TronMnemonic       string
	TronPassphrase     string
	EthereumMnemonic   string
	EthereumPassphrase string
	BSCMnemonic        string
	BSCPassphrase      string
	BitcoinMnemonic    string
	BitcoinPassphrase  string
	EthereumNetwork    models.EthereumNetwork
	TronNetwork        models.TronNetwork
	BSCNetwork         models.BSCNetwork
	BitcoinNetwork     models.BitcoinNetwork
	BitcoinAddressType address.BitcoinAddressType
	SolanaMnemonic     string
	SolanaPassphrase   string
	SolanaNetwork      models.SolanaNetwork
	XRPLMnemonic       string
	XRPLPassphrase     string
	XRPLNetwork        models.XRPLNetwork
}

func LoadConfig() (*Config, error) {
	var c EnvConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, err
	}
	EthereumNetwork := models.EthereumTestnetSepolia
	TronNetwork := models.TronTestnetNile
	BSCNetwork := models.BSCTestnet
	BitcoinNetwork := models.BitcoinSignet
	SolanaNetwork := models.SolanaDevnet
	BitcoinAddressType := address.P2TR
	XRPLNetwork := models.XRPLTestnet
	switch strings.ToLower(c.BitcoinAddressType) {
	case "p2tr":
		BitcoinAddressType = address.P2TR
	case "p2sh":
		BitcoinAddressType = address.P2SH
	case "p2wpkh":
		BitcoinAddressType = address.P2WPKH
	case "p2pkh":
		BitcoinAddressType = address.P2PKH
	default:
		return nil, fmt.Errorf("invalid bitcoin address type: %s. Available types: p2tr, p2sh, p2wpkh, p2pkh", c.BitcoinAddressType)
	}
	if c.Mode == ModeMainnet {
		EthereumNetwork = models.EthereumMainnet
		TronNetwork = models.TronMainnet
		BSCNetwork = models.BSCMainnet
		BitcoinNetwork = models.BitcoinMainnet
		SolanaNetwork = models.SolanaMainnet
		XRPLNetwork = models.XRPLMainnet
	}
	return &Config{
		Port:               c.Port,
		IJSONEndpoint:      c.IJSONEndpoint,
		DatabaseDSN:        c.DatabaseDSN,
		TronMnemonic:       c.TronMnemonic,
		TronPassphrase:     c.TronPassphrase,
		EthereumMnemonic:   c.EthereumMnemonic,
		EthereumPassphrase: c.EthereumPassphrase,
		BSCMnemonic:        c.BSCMnemonic,
		BSCPassphrase:      c.BSCPassphrase,
		EthereumNetwork:    EthereumNetwork,
		TronNetwork:        TronNetwork,
		BSCNetwork:         BSCNetwork,
		BitcoinMnemonic:    c.BitcoinMnemonic,
		BitcoinPassphrase:  c.BitcoinPassphrase,
		BitcoinNetwork:     BitcoinNetwork,
		BitcoinAddressType: BitcoinAddressType,
		SolanaMnemonic:     c.SolanaMnemonic,
		SolanaPassphrase:   c.SolanaPassphrase,
		SolanaNetwork:      SolanaNetwork,
		XRPLMnemonic:       c.XRPLMnemonic,
		XRPLPassphrase:     c.XRPLPassphrase,
		XRPLNetwork:        XRPLNetwork,
	}, nil
}
