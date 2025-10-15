package utils

// m/44'/501'/account'/change'/index'
func SolanaPurpose() uint32               { return HardenedIndex(44) }
func SolanaCoinTypeHardened() uint32      { return HardenedIndex(SolanaCoinType) }
func SolanaAccount(account uint32) uint32 { return HardenedIndex(account) }
func SolanaChange(change uint32) uint32   { return HardenedIndex(change) }
func SolanaAddressIndex(i uint32) uint32  { return HardenedIndex(i) }
