package blockchain

import (
	"fiatless/internal/blockchain/ethereum"
	"fiatless/internal/blockchain/handler"

	bitcoin_handler "fiatless/internal/blockchain/bitcoin"
	bsc_handler "fiatless/internal/blockchain/bsc"
	solana_handler "fiatless/internal/blockchain/solana"
	tron_handler "fiatless/internal/blockchain/tron"
	xrpl_handler "fiatless/internal/blockchain/xrpl"
	"fiatless/pkg/bitcoin"
	bsc_client "fiatless/pkg/bsc/client"
	ethereum_client "fiatless/pkg/ethereum/client"
	solana_client "fiatless/pkg/solana"
	"fiatless/pkg/tron"

	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func CreateHandlers(
	ethereumClient *ethereum_client.EthereumClient,
	tron *tron.Tron,
	bsc *bsc_client.BSCClient,
	bitcoin *bitcoin.Bitcoin,
	solana *solana_client.Solana,
	xrpl *xrpl_rpc.Client,
) []handler.CommandHandler {
	handlers := []handler.CommandHandler{
		ethereum.NewBalanceHandler(ethereumClient),
		ethereum.NewWithdrawHandler(ethereumClient),

		tron_handler.NewBalanceHandler(tron),
		tron_handler.NewWithdrawHandler(tron),
		tron_handler.NewBlockParseHandler(tron),

		bsc_handler.NewBalanceHandler(bsc),
		bsc_handler.NewWithdrawHandler(bsc),

		bitcoin_handler.NewBalanceHandler(bitcoin),
		bitcoin_handler.NewBlockParseHandler(bitcoin),
		bitcoin_handler.NewWithdrawHandler(bitcoin),

		solana_handler.NewBalanceHandler(solana),
		solana_handler.NewWithdrawHandler(solana),
		solana_handler.NewBlockParseHandler(solana),
		solana_handler.NewTransactionsHandler(solana),

		xrpl_handler.NewBalanceHandler(xrpl),
		xrpl_handler.NewWithdrawHandler(xrpl),
	}
	return handlers
}
