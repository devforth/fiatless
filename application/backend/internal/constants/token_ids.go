package constants

import "fiatless/internal/models"

// Token IDs
const (
	// Native tokens
	TronMainnetNativeTokenID     = "ae0fc5ba-622e-4618-95bb-0bf3af7abc67"
	TronNileNativeTokenID        = "5062799b-7dc6-479e-a955-7f490e5b4606"
	EthereumMainnetNativeTokenID = "f9187174-5719-47db-98e5-68d6439bfc3b"
	EthereumSepoliaNativeTokenID = "e99b989b-151a-40ec-9abd-2dd4ef31ab07"
	BSCMainnetNativeTokenID      = "092958bf-8d94-4c65-b42a-4bb6e1f58dc6"
	BSCTestnetNativeTokenID      = "935562bf-c60b-43bc-925a-b176fbd54983"
	BitcoinMainnetNativeTokenID  = "d63b44eb-1891-4da4-b590-f6a896e0cdcc"
	BitcoinSignetNativeTokenID   = "b3fedd7e-afe7-4acb-ba2b-a71ace8babf5"
	SolanaMainnetNativeTokenID   = "0be8641e-77b7-4d69-9abf-5a450c85fea5"
	SolanaDevnetNativeTokenID    = "c3aa2f97-3ac1-4308-abc3-d68cf7e799d0"
	XRPLMainnetNativeTokenID     = "d1d4e1af-2d70-4468-9001-2cc392b2a53d"
	XRPLTestnetNativeTokenID     = "e2fff7be-7acc-44bb-a366-a3011d6d83f5"

	// Contract tokens
	USDTTronMainnetTokenID     = "db9da771-2e0b-468d-b502-47eeb9260b68"
	USDTEthereumMainnetTokenID = "ad1028a9-3858-4bc0-8aac-0cc83865cd9b"
	USDTBSCMainnetTokenID      = "8e1f765d-718b-4ea1-be24-afb5e02a1e40"
	USDTBSCTestnetTokenID      = "d58bb5f2-1dbb-4507-b821-b89d9065fe61"
	USDCEthereumMainnetTokenID = "27626754-af8d-4bd7-861d-1e48410f3102"
	USDCEthereumSepoliaTokenID = "ab62563a-9868-407b-b19d-e78436f28382"
)

func GetBitcoinTokenID(network models.BitcoinNetwork) string {
	switch network {
	case models.BitcoinMainnet:
		return BitcoinMainnetNativeTokenID
	case models.BitcoinSignet:
		return BitcoinSignetNativeTokenID
	}
	return BitcoinMainnetNativeTokenID
}

func GetTRXTokenID(network models.TronNetwork) string {
	switch network {
	case models.TronMainnet:
		return TronMainnetNativeTokenID
	case models.TronTestnetNile:
		return TronNileNativeTokenID
	}
	return TronMainnetNativeTokenID
}

func GetSolanaTokenID(network models.SolanaNetwork) string {
	switch network {
	case models.SolanaMainnet:
		return SolanaMainnetNativeTokenID
	case models.SolanaDevnet:
		return SolanaDevnetNativeTokenID
	}
	return SolanaMainnetNativeTokenID
}

func GetXRPLTokenID(network models.XRPLNetwork) string {
	switch network {
	case models.XRPLMainnet:
		return XRPLMainnetNativeTokenID
	case models.XRPLTestnet:
		return XRPLTestnetNativeTokenID
	}
	return XRPLMainnetNativeTokenID
}
