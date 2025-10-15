package constants

import (
	"fiatless/internal/models"

	"github.com/google/uuid"
)

// Blockchain IDs
const (
	// Tron networks
	TronMainnetBlockchainID = "2f3e352d-6f53-481d-80d4-409d2dcbf4f9"
	TronNileBlockchainID    = "ad16e7b1-48e9-4439-97df-6a0d0d71cfeb"

	// BSC networks
	BSCMainnetBlockchainID = "b8aec6f2-efb1-4e4c-8657-4bc81b4f79e7"
	BSCTestnetBlockchainID = "88304b08-99ad-494b-b5c8-08b21ae820ae"

	// Ethereum networks
	EthereumSepoliaBlockchainID = "6e0b6d56-aada-4742-9834-e30fa25e13ce"
	EthereumMainnetBlockchainID = "e80dd2cc-94aa-44c3-9af4-7aaee64d1ab6"

	// Bitcoin networks
	BitcoinMainnetBlockchainID = "a1bfa33e-faad-48b3-9342-d7d99034827e"
	BitcoinSignetBlockchainID  = "3a6a1d9d-e8ec-4808-a296-d6be30c8fdd3"

	// Solana networks
	SolanaMainnetBlockchainID = "50309da1-402e-46fb-b320-a7bf58034cd2"
	SolanaDevnetBlockchainID  = "b2a6f906-1619-47b5-9c02-97f03e35a0c3"

	// XRPL networks
	XRPLMainnetBlockchainID = "714ba0a6-0704-4d95-9f64-3bf0ca3530dd"
	XRPLTestnetBlockchainID = "7c0406c8-5c50-4341-a9bb-48b0de5c55ab"
)

func GetEthereumBlockchainID(network models.EthereumNetwork) uuid.UUID {
	switch network {
	case models.EthereumMainnet:
		return uuid.MustParse(EthereumMainnetBlockchainID)
	case models.EthereumTestnetSepolia:
		return uuid.MustParse(EthereumSepoliaBlockchainID)
	}
	return uuid.MustParse(EthereumMainnetBlockchainID)
}

func GetTronBlockchainID(network models.TronNetwork) uuid.UUID {
	switch network {
	case models.TronMainnet:
		return uuid.MustParse(TronMainnetBlockchainID)
	case models.TronTestnetNile:
		return uuid.MustParse(TronNileBlockchainID)
	}
	return uuid.MustParse(TronMainnetBlockchainID)
}

func GetBSCBlockchainID(network models.BSCNetwork) uuid.UUID {
	switch network {
	case models.BSCMainnet:
		return uuid.MustParse(BSCMainnetBlockchainID)
	case models.BSCTestnet:
		return uuid.MustParse(BSCTestnetBlockchainID)
	}
	return uuid.MustParse(BSCMainnetBlockchainID)
}

func GetBitcoinBlockchainID(network models.BitcoinNetwork) uuid.UUID {
	switch network {
	case models.BitcoinMainnet:
		return uuid.MustParse(BitcoinMainnetBlockchainID)
	case models.BitcoinSignet:
		return uuid.MustParse(BitcoinSignetBlockchainID)
	}
	return uuid.MustParse(BitcoinMainnetBlockchainID)
}

func GetSolanaBlockchainID(network models.SolanaNetwork) uuid.UUID {
	switch network {
	case models.SolanaMainnet:
		return uuid.MustParse(SolanaMainnetBlockchainID)
	case models.SolanaDevnet:
		return uuid.MustParse(SolanaDevnetBlockchainID)
	}
	return uuid.MustParse(SolanaMainnetBlockchainID)
}

func GetSolanaNetByBlockchainID(blockchainID uuid.UUID) models.SolanaNetwork {
	switch blockchainID {
	case uuid.MustParse(SolanaMainnetBlockchainID):
		return models.SolanaMainnet
	case uuid.MustParse(SolanaDevnetBlockchainID):
		return models.SolanaDevnet
	}
	return models.SolanaMainnet
}

func GetBitcoinNetByBlockchainID(blockchainID uuid.UUID) models.BitcoinNetwork {
	switch blockchainID {
	case uuid.MustParse(BitcoinMainnetBlockchainID):
		return models.BitcoinMainnet
	case uuid.MustParse(BitcoinSignetBlockchainID):
		return models.BitcoinSignet
	}
	return models.BitcoinMainnet
}

func GetXRPLBlockchainID(network models.XRPLNetwork) uuid.UUID {
	switch network {
	case models.XRPLMainnet:
		return uuid.MustParse(XRPLMainnetBlockchainID)
	case models.XRPLTestnet:
		return uuid.MustParse(XRPLTestnetBlockchainID)
	}
	return uuid.MustParse(XRPLMainnetBlockchainID)
}
