package main

import (
	"context"
	"fiatless/cmd/payment-processor/config"
	"fiatless/internal/constants"
	"fiatless/internal/db"
	"fiatless/internal/ijson"
	"fiatless/internal/repositories"
	"fiatless/internal/routes"
	bitcoin_routes "fiatless/internal/routes/bitcoin"
	bsc_routes "fiatless/internal/routes/bsc"
	ethereum_routes "fiatless/internal/routes/ethereum"
	solana_routes "fiatless/internal/routes/solana"
	tron_routes "fiatless/internal/routes/tron"
	xrpl_routes "fiatless/internal/routes/xrpl"
	"fiatless/internal/services"
	"fiatless/internal/tasks"
	"fiatless/internal/vars"
	"fiatless/pkg/bitcoin"
	btc_address "fiatless/pkg/bitcoin/address"
	"fiatless/pkg/bsc"
	"fiatless/pkg/ethereum"
	"fiatless/pkg/solana"
	"fiatless/pkg/tron"
	"fiatless/pkg/wallet"
	"fiatless/pkg/xrpl"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ggicci/httpin"
	httpin_integration "github.com/ggicci/httpin/integration"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func initDatabaseVariables(repositories *repositories.Repositories) {
	ctx := context.Background()

	if err := constants.InitDatabaseData(ctx, repositories); err != nil {
		log.Fatalf("Failed to initialize database data: %v", err)
	}

	log.Println("Database initialization completed successfully")
}

func init() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	vars.Config = config
	db := db.Connect(config.DatabaseDSN)
	httpin_integration.UseGorillaMux("path", mux.Vars)

	// init vars
	IJSONClient := ijson.NewIJSONClient(config.IJSONEndpoint)
	tronService := services.NewServiceFactory(IJSONClient).CreateTronService()
	ethereumService := services.NewServiceFactory(IJSONClient).CreateEthereumService()
	bscService := services.NewServiceFactory(IJSONClient).CreateBSCService()
	bitcoinService := services.NewServiceFactory(IJSONClient).CreateBitcoinService()
	solanaService := services.NewServiceFactory(IJSONClient).CreateSolanaService()
	xrplService := services.NewServiceFactory(IJSONClient).CreateXRPLService()
	repositories := repositories.NewRepositories(db)
	vars.Repositories = repositories
	initDatabaseVariables(repositories)
	if config.EthereumMnemonic != "" {
		vars.EthereumWalletManager = ethereum.NewWalletManager(wallet.NewBIP44KeyGenerator(config.EthereumMnemonic, config.EthereumPassphrase), ethereumService, repositories, config.EthereumNetwork)
		err := vars.EthereumWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize ethereum wallet manager: %v", err)
		}
	}
	if config.TronMnemonic != "" {
		vars.TronWalletManager = tron.NewWalletManager(wallet.NewBIP44KeyGenerator(config.TronMnemonic, config.TronPassphrase), tronService, repositories, config.TronNetwork)
		err := vars.TronWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize tron wallet manager: %v", err)
		}
	}
	if config.BSCMnemonic != "" {
		vars.BSCWalletManager = bsc.NewWalletManager(wallet.NewBIP44KeyGenerator(config.BSCMnemonic, config.BSCPassphrase), bscService, repositories, config.BSCNetwork)
		err := vars.BSCWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize bsc wallet manager: %v", err)
		}
	}
	if config.SolanaMnemonic != "" {
		solKeyGen := wallet.NewSLIP10Ed25519KeyGenerator(config.SolanaMnemonic, config.SolanaPassphrase)
		vars.SolanaWalletManager = solana.NewWalletManager(solKeyGen, solanaService, repositories, config.SolanaNetwork)
		err := vars.SolanaWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize solana wallet manager: %v", err)
		}
	}
	if config.XRPLMnemonic != "" {
		vars.XRPLWalletManager = xrpl.NewWalletManager(wallet.NewBIP44KeyGenerator(config.XRPLMnemonic, config.XRPLPassphrase), xrplService, repositories, config.XRPLNetwork)
		err := vars.XRPLWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize xrpl wallet manager: %v", err)
		}
	}
	if config.BitcoinMnemonic != "" {
		// Select generator by address type: P2TR -> BIP86, else default to BIP84
		var keyGen wallet.KeyGenerator
		if config.BitcoinAddressType == btc_address.P2TR {
			keyGen = wallet.NewBIP86KeyGenerator(config.BitcoinMnemonic, config.BitcoinPassphrase)
		} else {
			keyGen = wallet.NewBIP84KeyGenerator(config.BitcoinMnemonic, config.BitcoinPassphrase)
		}
		vars.BitcoinWalletManager = bitcoin.NewWalletManager(keyGen, bitcoinService, repositories, config.BitcoinNetwork, config.BitcoinAddressType)
		err := vars.BitcoinWalletManager.Init()
		if err != nil {
			log.Fatalf("Failed to initialize bitcoin wallet manager: %v", err)
		}
	}
	vars.TaskFactory = tasks.NewTaskFactory(tronService, bitcoinService, solanaService, xrplService)
	vars.TaskManager = tasks.NewTaskManager()
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	router := mux.NewRouter()

	staticFileDirectory := http.Dir("./static/")
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))

	router.PathPrefix("/static/").Handler(staticFileHandler).Methods("GET")

	if vars.TronWalletManager != nil {
		sTron := router.PathPrefix("/api/v1/tron").Subrouter()
		sTron.Handle("/create-wallet", alice.New(httpin.NewInput(tron_routes.CreateWalletRequest{})).ThenFunc(tron_routes.CreateWallet)).Methods("POST")
		sTron.Handle("/{address}/balance", alice.New(httpin.NewInput(tron_routes.GetBalanceRequest{})).ThenFunc(tron_routes.GetBalance)).Methods("GET")
		sTron.Handle("/withdraw", alice.New(httpin.NewInput(tron_routes.WithdrawRequest{})).ThenFunc(tron_routes.Withdraw)).Methods("POST")
		sTron.Handle("/start-sweeping", alice.New(httpin.NewInput(tron_routes.StartSweepingRequest{})).ThenFunc(tron_routes.StartSweeping)).Methods("POST")
		sTron.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(tron_routes.GetMainWalletAddress)).Methods("GET")
		sTron.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(tron_routes.GetTronTokens)).Methods("GET")
	} else {
		log.Println("Tron wallet manager is not initialized")
	}

	if vars.EthereumWalletManager != nil {
		sEthereum := router.PathPrefix("/api/v1/ethereum").Subrouter()
		sEthereum.Handle("/create-wallet", alice.New(httpin.NewInput(ethereum_routes.CreateWalletRequest{})).ThenFunc(ethereum_routes.CreateWallet)).Methods("POST")
		sEthereum.Handle("/{address}/balance", alice.New(httpin.NewInput(ethereum_routes.GetBalanceRequest{})).ThenFunc(ethereum_routes.GetBalance)).Methods("GET")
		sEthereum.Handle("/withdraw", alice.New(httpin.NewInput(ethereum_routes.WithdrawRequest{})).ThenFunc(ethereum_routes.Withdraw)).Methods("POST")
		sEthereum.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(ethereum_routes.GetEthereumTokens)).Methods("GET")
		sEthereum.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(ethereum_routes.GetMainWalletAddress)).Methods("GET")
	} else {
		log.Println("Ethereum wallet manager is not initialized")
	}

	if vars.BSCWalletManager != nil {
		sBSC := router.PathPrefix("/api/v1/bsc").Subrouter()
		sBSC.Handle("/create-wallet", alice.New(httpin.NewInput(bsc_routes.CreateWalletRequest{})).ThenFunc(bsc_routes.CreateWallet)).Methods("POST")
		sBSC.Handle("/{address}/balance", alice.New(httpin.NewInput(bsc_routes.GetBalanceRequest{})).ThenFunc(bsc_routes.GetBalance)).Methods("GET")
		sBSC.Handle("/withdraw", alice.New(httpin.NewInput(bsc_routes.WithdrawRequest{})).ThenFunc(bsc_routes.Withdraw)).Methods("POST")
		sBSC.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(bsc_routes.GetBSCTokens)).Methods("GET")
		sBSC.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(bsc_routes.GetMainWalletAddress)).Methods("GET")
	} else {
		log.Println("BSC wallet manager is not initialized")
	}

	if vars.SolanaWalletManager != nil {
		sSolana := router.PathPrefix("/api/v1/solana").Subrouter()
		sSolana.Handle("/create-wallet", alice.New(httpin.NewInput(solana_routes.CreateWalletRequest{})).ThenFunc(solana_routes.CreateWallet)).Methods("POST")
		sSolana.Handle("/{address}/balance", alice.New(httpin.NewInput(solana_routes.GetBalanceRequest{})).ThenFunc(solana_routes.GetBalance)).Methods("GET")
		sSolana.Handle("/withdraw", alice.New(httpin.NewInput(solana_routes.WithdrawRequest{})).ThenFunc(solana_routes.Withdraw)).Methods("POST")
		sSolana.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(solana_routes.GetSolanaTokens)).Methods("GET")
		sSolana.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(solana_routes.GetMainWalletAddress)).Methods("GET")
	} else {
		log.Println("Solana wallet manager is not initialized")
	}

	if vars.BitcoinWalletManager != nil {
		sBitcoin := router.PathPrefix("/api/v1/bitcoin").Subrouter()
		sBitcoin.Handle("/{address}/balance", alice.New(httpin.NewInput(bitcoin_routes.GetBalanceRequest{})).ThenFunc(bitcoin_routes.GetBalance)).Methods("GET")
		sBitcoin.Handle("/create-wallet", alice.New(httpin.NewInput(bitcoin_routes.CreateWalletRequest{})).ThenFunc(bitcoin_routes.CreateWallet)).Methods("POST")
		sBitcoin.Handle("/withdraw", alice.New(httpin.NewInput(bitcoin_routes.WithdrawRequest{})).ThenFunc(bitcoin_routes.Withdraw)).Methods("POST")
		sBitcoin.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(bitcoin_routes.GetBitcoinTokens)).Methods("GET")
		sBitcoin.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(bitcoin_routes.GetMainWalletAddress)).Methods("GET")
	} else {
		log.Println("Bitcoin wallet manager is not initialized")
	}

	if vars.XRPLWalletManager != nil {
		sXRPL := router.PathPrefix("/api/v1/xrpl").Subrouter()
		sXRPL.Handle("/{address}/balance", alice.New(httpin.NewInput(xrpl_routes.GetBalanceRequest{})).ThenFunc(xrpl_routes.GetBalance)).Methods("GET")
		sXRPL.Handle("/create-wallet", alice.New(httpin.NewInput(xrpl_routes.CreateWalletRequest{})).ThenFunc(xrpl_routes.CreateWallet)).Methods("POST")
		sXRPL.Handle("/withdraw", alice.New(httpin.NewInput(xrpl_routes.WithdrawRequest{})).ThenFunc(xrpl_routes.Withdraw)).Methods("POST")
		sXRPL.Handle("/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(xrpl_routes.GetXRPLTokens)).Methods("GET")
		sXRPL.Handle("/main-wallet/address", alice.New(httpin.NewInput(routes.GetMainWalletAddressRequest{})).ThenFunc(xrpl_routes.GetMainWalletAddress)).Methods("GET")
	} else {
		log.Println("XRPL wallet manager is not initialized")
	}

	router.Handle("/api/v1/tokens", alice.New(httpin.NewInput(routes.GetTokensRequest{})).ThenFunc(routes.GetTokens)).Methods("GET")
	router.Handle("/api/v1/tokens/{token_id}", alice.New(httpin.NewInput(routes.GetTokenRequest{})).ThenFunc(routes.GetToken)).Methods("GET")
	fmt.Printf("Server started on port %s\n", config.Port)
	go http.ListenAndServe(":"+config.Port, router)
	if vars.SolanaWalletManager != nil {
		vars.TaskManager.Add(vars.TaskFactory.CreateSolanaTransactionsTask(constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork), 12*time.Hour, 30*time.Second, vars.SolanaWalletManager, vars.Repositories))
	}
	vars.TaskManager.Run()
	// if vars.TronWalletManager != nil {

	// 	// vars.TaskManager.Add(vars.TaskFactory.CreateTronBlocksTask(constants.GetTronBlockchainID(vars.Config.TronNetwork), 30*time.Second, 3*time.Second, vars.TronWalletManager, vars.Repositories))
	// 	// vars.TaskManager.Add(vars.TaskFactory.CreateBitcoinBlocksTask(constants.GetBitcoinBlockchainID(vars.Config.BitcoinNetwork), 50*time.Second, 1*time.Minute, vars.BitcoinWalletManager, vars.Repositories))
	// 	// vars.TaskManager.Add(vars.TaskFactory.CreateSolanaBlocksTask(constants.GetSolanaBlockchainID(vars.Config.SolanaNetwork), 50*time.Second, 1*time.Minute, vars.SolanaWalletManager, vars.Repositories))
	// } else {
	// 	http.ListenAndServe(":"+config.Port, router)
	// }
}
