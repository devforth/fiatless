package constants

import (
	"context"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InitDatabaseData initializes all blockchain and token data using code instead of SQL
func InitDatabaseData(ctx context.Context, repos *repositories.Repositories) error {
	// Initialize blockchains first
	if err := initBlockchains(ctx, repos); err != nil {
		return err
	}

	// Initialize tokens
	if err := initTokens(ctx, repos); err != nil {
		return err
	}

	return nil
}

// initBlockchains creates all blockchain records
func initBlockchains(ctx context.Context, repos *repositories.Repositories) error {
	blockchains := []models.Blockchain{
		{
			ID:          uuid.MustParse(TronMainnetBlockchainID),
			Name:        "Tron",
			Symbol:      "TRX",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://tronscan.org/#/transaction/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(TronNileBlockchainID),
			Name:        "Tron",
			Symbol:      "TRX",
			Network:     "NILE",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://nile.tronscan.org/#/transaction/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(EthereumSepoliaBlockchainID),
			Name:        "Ethereum",
			Symbol:      "ETH",
			Network:     "SEPOLIA",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://sepolia.etherscan.io/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(EthereumMainnetBlockchainID),
			Name:        "Ethereum",
			Symbol:      "ETH",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://etherscan.io/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(BSCTestnetBlockchainID),
			Name:        "Binance Smart Chain",
			Symbol:      "BSC",
			Network:     "TESTNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://testnet.bscscan.com/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(BSCMainnetBlockchainID),
			Name:        "Binance Smart Chain",
			Symbol:      "BSC",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://bscscan.com/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(BitcoinMainnetBlockchainID),
			Name:        "Bitcoin",
			Symbol:      "BTC",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://mempool.space/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(BitcoinSignetBlockchainID),
			Name:        "Bitcoin",
			Symbol:      "BTC",
			Network:     "SIGNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://mempool.space/signet/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(SolanaMainnetBlockchainID),
			Name:        "Solana",
			Symbol:      "SOL",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://solscan.io/tx/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(SolanaDevnetBlockchainID),
			Name:        "Solana",
			Symbol:      "SOL",
			Network:     "DEVNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://solscan.io/tx/{{tx_id}}?cluster=devnet",
		},
		{
			ID:          uuid.MustParse(XRPLMainnetBlockchainID),
			Name:        "XRPL",
			Symbol:      "XRP",
			Network:     "MAINNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://livenet.xrpl.org/transactions/{{tx_id}}",
		},
		{
			ID:          uuid.MustParse(XRPLTestnetBlockchainID),
			Name:        "XRPL",
			Symbol:      "XRP",
			Network:     "TESTNET",
			IsActive:    true,
			LogoURL:     "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			ExplorerURL: "https://testnet.xrpl.org/transactions/{{tx_id}}",
		},
	}

	for _, blockchain := range blockchains {
		// Check if blockchain already exists
		existing, err := repos.Blockchain.GetBlockchainByID(ctx, blockchain.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if existing == nil {
			if err := repos.Blockchain.AddBlockchain(ctx, &blockchain); err != nil {
				return err
			}
		}
	}

	return nil
}

// initTokens creates all token records
func initTokens(ctx context.Context, repos *repositories.Repositories) error {
	tokens := []models.Token{
		// Native tokens
		{
			ID:           uuid.MustParse(TronMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "Tron",
			Symbol:       "TRX",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(TronMainnetBlockchainID),
			IsActive:     false,
			LogoURL:      "/static/trx.svg",
			YahooSymbol:  "TRX-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(TronNileNativeTokenID),
			TokenID:      nil,
			Name:         "Tron",
			Symbol:       "TRX",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(TronNileBlockchainID),
			IsActive:     false,
			LogoURL:      "/static/trx.svg",
			YahooSymbol:  "TRX-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(EthereumMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "Ethereum",
			Symbol:       "ETH",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(EthereumMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/eth.svg",
			YahooSymbol:  "ETH-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(EthereumSepoliaNativeTokenID),
			TokenID:      nil,
			Name:         "Ethereum",
			Symbol:       "ETH",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(EthereumSepoliaBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/eth.svg",
			YahooSymbol:  "ETH-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(BSCMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "Binance Smart Chain",
			Symbol:       "BNB",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(BSCMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/bnb.svg",
			YahooSymbol:  "BNB-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(BSCTestnetNativeTokenID),
			TokenID:      nil,
			Name:         "Binance Smart Chain",
			Symbol:       "BNB",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(BSCTestnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/bnb.svg",
			YahooSymbol:  "BNB-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(BitcoinMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "Bitcoin",
			Symbol:       "BTC",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(BitcoinMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/btc.svg",
			YahooSymbol:  "BTC-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(BitcoinSignetNativeTokenID),
			TokenID:      nil,
			Name:         "Bitcoin",
			Symbol:       "BTC",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(BitcoinSignetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/btc.svg",
			YahooSymbol:  "BTC-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(XRPLMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "XRPL",
			Symbol:       "XRP",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(XRPLMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/xrp.svg",
			YahooSymbol:  "XRP-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(XRPLTestnetNativeTokenID),
			TokenID:      nil,
			Name:         "XRPL",
			Symbol:       "XRP",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(XRPLTestnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/xrp.svg",
			YahooSymbol:  "XRP-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(SolanaMainnetNativeTokenID),
			TokenID:      nil,
			Name:         "Solana",
			Symbol:       "SOL",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(SolanaMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/sol.svg",
			YahooSymbol:  "SOL-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(SolanaDevnetNativeTokenID),
			TokenID:      nil,
			Name:         "Solana",
			Symbol:       "SOL",
			Type:         models.TokenTypeNative,
			BlockchainID: uuid.MustParse(SolanaDevnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/sol.svg",
			YahooSymbol:  "SOL-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Contract tokens
		{
			ID:           uuid.MustParse(USDTTronMainnetTokenID),
			TokenID:      &USDTTronMainnetContractAddress,
			Name:         "USD Tether",
			Symbol:       "USDT",
			Type:         models.TokenTypeTRC20,
			BlockchainID: uuid.MustParse(TronMainnetBlockchainID),
			IsActive:     false,
			LogoURL:      "/static/usdt.svg",
			YahooSymbol:  "USDT-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(USDTEthereumMainnetTokenID),
			TokenID:      &USDTEthereumMainnetContractAddress,
			Name:         "USD Tether",
			Symbol:       "USDT",
			Type:         models.TokenTypeERC20,
			BlockchainID: uuid.MustParse(EthereumMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/usdt.svg",
			YahooSymbol:  "USDT-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(USDTBSCMainnetTokenID),
			TokenID:      &USDTBSCMainnetContractAddress,
			Name:         "USD Tether",
			Symbol:       "USDT",
			Type:         models.TokenTypeBEP20,
			BlockchainID: uuid.MustParse(BSCMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/usdt.svg",
			YahooSymbol:  "USDT-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(USDTBSCTestnetTokenID),
			TokenID:      &USDTBSCTestnetContractAddress,
			Name:         "USD Tether",
			Symbol:       "USDT",
			Type:         models.TokenTypeBEP20,
			BlockchainID: uuid.MustParse(BSCTestnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/usdt.svg",
			YahooSymbol:  "USDT-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(USDCEthereumMainnetTokenID),
			TokenID:      &USDCEthereumMainnetContractAddress,
			Name:         "USD Coin",
			Symbol:       "USDC",
			Type:         models.TokenTypeERC20,
			BlockchainID: uuid.MustParse(EthereumMainnetBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/usdc.svg",
			YahooSymbol:  "USDC-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse(USDCEthereumSepoliaTokenID),
			TokenID:      &USDCEthereumSepoliaContractAddress,
			Name:         "USD Coin",
			Symbol:       "USDC",
			Type:         models.TokenTypeERC20,
			BlockchainID: uuid.MustParse(EthereumSepoliaBlockchainID),
			IsActive:     true,
			LogoURL:      "/static/usdc.svg",
			YahooSymbol:  "USDC-USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, token := range tokens {
		// Check if token already exists
		existing, err := repos.Token.GetToken(ctx, &models.Token{ID: token.ID})
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if existing == nil {
			if err := repos.Token.AddToken(ctx, &token); err != nil {
				return err
			}
		}
	}

	return nil
}
