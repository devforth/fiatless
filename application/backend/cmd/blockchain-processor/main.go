package main

import (
	"fiatless/cmd/blockchain-processor/config"
	"fiatless/cmd/blockchain-processor/processor"
	"fiatless/internal/blockchain"
	"fiatless/internal/ijson"
	"fiatless/pkg/bitcoin"
	bitcoin_client "fiatless/pkg/bitcoin/client"
	bsc_client "fiatless/pkg/bsc/client"
	ethereum_client "fiatless/pkg/ethereum/client"
	"fiatless/pkg/httpclient"
	solana "fiatless/pkg/solana"
	"fiatless/pkg/tron"
	tron_client "fiatless/pkg/tron/client/grpc"
	"log"
	"net/http"

	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
	solana_client "github.com/gagliardetto/solana-go/rpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	tronClient, err := tron_client.NewTronClient(cfg.TronGRPCEndpoint, cfg.TronRPS/float64(cfg.WorkerCount))
	if err != nil {
		log.Fatalf("Failed to create Tron client: %v", err)
	}

	tron := tron.NewTron(tronClient)

	ethereumClient, err := ethereum_client.NewEthereumClient(cfg.EthereumRPCEndpoint, cfg.EthereumRPS/float64(cfg.WorkerCount))
	if err != nil {
		log.Fatalf("Failed to create Ethereum client: %v", err)
	}

	bscClient, err := bsc_client.NewBSCClient(cfg.BSCRPCEndpoint, cfg.BSCRPS/float64(cfg.WorkerCount))
	if err != nil {
		log.Fatalf("Failed to create BSC client: %v", err)
	}

	ijsonClient := ijson.NewIJSONClient(cfg.IJSONEndpoint)

	commandProcessor := processor.NewCommandProcessor(ijsonClient)

	bitcoinClient, err := bitcoin_client.NewBitcoinClient(cfg.BitcoinRPCEndpoint, cfg.BitcoinRPS/float64(cfg.WorkerCount))
	if err != nil {
		log.Fatalf("Failed to create Bitcoin client: %v", err)
	}

	bitcoin := bitcoin.NewBitcoin(bitcoinClient)

	solanaClient := solana_client.NewWithCustomRPCClient(solana_client.NewWithRateLimit(cfg.SolanaRPCEndpoint, int(cfg.SolanaRPS/float64(cfg.WorkerCount))))
	solana := solana.NewSolana(solanaClient)

	xrplClient := xrpl_rpc.NewClient(&xrpl_rpc.Config{URL: cfg.XRPLRPCEndpoint, HTTPClient: &http.Client{Transport: httpclient.NewRateLimitedTransport(cfg.XRPLRPS/float64(cfg.WorkerCount), nil)}})

	handlers := blockchain.CreateHandlers(ethereumClient, tron, bscClient, bitcoin, solana, xrplClient)
	commandProcessor.RegisterHandlers(handlers)

	log.Println("Blockchain processor started, waiting for commands...")
	commandProcessor.Start()
}
