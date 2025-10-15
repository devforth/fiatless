package config

import (
	"context"
	"fiatless/internal/models"

	"github.com/sethvargo/go-envconfig"
)

type Mode string

const (
	ModeTestnet Mode = "testnet"
	ModeMainnet Mode = "mainnet"
)

type TronNode string

const (
	TronNodeTrongrid TronNode = "trongrid"
)

type EthereumNode string

const (
	EthereumNodePublicnode EthereumNode = "publicnode"
)

type EnvConfig struct {
	Mode         Mode         `env:"MODE, default=testnet"`
	WorkerCount  int          `env:"WORKER_COUNT, default=1"`
	TronRPS      float64      `env:"TRON_RPS, default=5"`
	TronNode     TronNode     `env:"TRON_NODE, default=trongrid"`
	EthereumNode EthereumNode `env:"ETHEREUM_NODE, default=publicnode"`
	EthereumRPS  float64      `env:"ETHEREUM_RPS, default=5"`
	BSCNode      string       `env:"BSC_RPC_ENDPOINT, default=bnbchain"`
	BSCRPS       float64      `env:"BSC_RPS, default=5"`
	BitcoinNode  string       `env:"BITCOIN_RPC_ENDPOINT, default=signet"`
	BitcoinRPS   float64      `env:"BITCOIN_RPS, default=5"`
	SolanaNode   string       `env:"SOLANA_RPC_ENDPOINT, default=solana"`
	SolanaRPS    float64      `env:"SOLANA_RPS, default=5"`
	XRPLRPS      float64      `env:"XRPL_RPS, default=5"`
}

const (
	TRON_MAINNET_GRPC_ENDPOINT    = "grpc.trongrid.io:50051"
	TRON_NILE_GRPC_ENDPOINT       = "grpc.nile.trongrid.io:50051"
	ETHEREUM_SEPOLIA_RPC_ENDPOINT = "https://ethereum-sepolia-rpc.publicnode.com"
	ETHEREUM_MAINNET_RPC_ENDPOINT = "https://ethereum-rpc.publicnode.com"
	IJSON_ENDPOINT                = "http://localhost:8001"
	BSC_MAINNET_RPC_ENDPOINT      = "https://bsc-dataseed.bnbchain.org"
	BSC_TESTNET_RPC_ENDPOINT      = "https://bsc-testnet.bnbchain.org"
	BITCOIN_MAINNET_RPC_ENDPOINT  = "https://bitcoin-mainnet.publicnode.com"
	BITCOIN_SIGNET_RPC_ENDPOINT   = "https://rpc.ankr.com/btc_signet/04409891e7f98f0a12ccf296e1be5b1d60a25b5757941285050f38d7c4bac788"
	SOLANA_MAINNET_RPC_ENDPOINT   = "https://api.mainnet-beta.solana.com"
	SOLANA_DEVNET_RPC_ENDPOINT    = "https://api.devnet.solana.com"
	XRPL_MAINNET_RPC_ENDPOINT     = "https://xrplcluster.com/"
	XRPL_TESTNET_RPC_ENDPOINT     = "https://s.altnet.rippletest.net:51234/"
)

type Config struct {
	TronGRPCEndpoint    string
	EthereumRPCEndpoint string
	IJSONEndpoint       string
	WorkerCount         int
	TronRPS             float64
	EthereumRPS         float64
	BSCRPCEndpoint      string
	BSCRPS              float64
	SolanaRPCEndpoint   string
	SolanaRPS           float64
	Port                string
	Mnemonic            string
	Passphrase          string
	EthereumNetwork     models.EthereumNetwork
	TronNetwork         models.TronNetwork
	BitcoinRPCEndpoint  string
	BitcoinRPS          float64
	Mode                Mode
	XRPLRPCEndpoint     string
	XRPLRPS             float64
}

func LoadConfig() (*Config, error) {
	var c EnvConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, err
	}

	tronGRPCEndpoint := TRON_MAINNET_GRPC_ENDPOINT
	if c.Mode == ModeTestnet {
		tronGRPCEndpoint = TRON_NILE_GRPC_ENDPOINT
	}

	ethereumRPCEndpoint := ETHEREUM_SEPOLIA_RPC_ENDPOINT
	if c.Mode == ModeMainnet {
		ethereumRPCEndpoint = ETHEREUM_MAINNET_RPC_ENDPOINT
	}

	bscRPCEndpoint := BSC_MAINNET_RPC_ENDPOINT
	if c.Mode == ModeTestnet {
		bscRPCEndpoint = BSC_TESTNET_RPC_ENDPOINT
	}

	bitcoinRPCEndpoint := BITCOIN_MAINNET_RPC_ENDPOINT
	if c.Mode == ModeTestnet {
		bitcoinRPCEndpoint = BITCOIN_SIGNET_RPC_ENDPOINT
	}

	solanaRPCEndpoint := SOLANA_MAINNET_RPC_ENDPOINT
	if c.Mode == ModeTestnet {
		solanaRPCEndpoint = SOLANA_DEVNET_RPC_ENDPOINT
	}

	xrplRPCEndpoint := XRPL_MAINNET_RPC_ENDPOINT
	if c.Mode == ModeTestnet {
		xrplRPCEndpoint = XRPL_TESTNET_RPC_ENDPOINT
	}

	return &Config{
		TronGRPCEndpoint:    tronGRPCEndpoint,
		EthereumRPCEndpoint: ethereumRPCEndpoint,
		IJSONEndpoint:       IJSON_ENDPOINT,
		WorkerCount:         c.WorkerCount,
		TronRPS:             c.TronRPS,
		EthereumRPS:         c.EthereumRPS,
		BSCRPCEndpoint:      bscRPCEndpoint,
		BSCRPS:              c.BSCRPS,
		BitcoinRPCEndpoint:  bitcoinRPCEndpoint,
		BitcoinRPS:          c.BitcoinRPS,
		SolanaRPCEndpoint:   solanaRPCEndpoint,
		SolanaRPS:           c.SolanaRPS,
		Mode:                c.Mode,
		XRPLRPCEndpoint:     xrplRPCEndpoint,
		XRPLRPS:             c.XRPLRPS,
	}, nil
}
