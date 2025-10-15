package vars

import (
	"fiatless/cmd/payment-processor/config"
	"fiatless/internal/repositories"
	"fiatless/internal/tasks"
	"fiatless/pkg/bitcoin"
	"fiatless/pkg/bsc"
	"fiatless/pkg/ethereum"
	"fiatless/pkg/solana"
	"fiatless/pkg/tron"
	"fiatless/pkg/xrpl"
)

var (
	TronWalletManager     *tron.WalletManager
	EthereumWalletManager *ethereum.WalletManager
	BitcoinWalletManager  *bitcoin.WalletManager
	BSCWalletManager      *bsc.WalletManager
	SolanaWalletManager   *solana.WalletManager
	XRPLWalletManager     *xrpl.WalletManager
	Repositories          *repositories.Repositories
	TaskManager           *tasks.TaskManager
	Config                *config.Config
	TaskFactory           *tasks.TaskFactory
)
