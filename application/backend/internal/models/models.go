package models

import (
	bsc_address "fiatless/pkg/bsc/address"
	ethereum_address "fiatless/pkg/ethereum/address"
	tron_address "fiatless/pkg/tron/address"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type TokenType string

const (
	TokenTypeTRC20  TokenType = "TRC-20"
	TokenTypeTRC10  TokenType = "TRC-10"
	TokenTypeNative TokenType = "NATIVE"
	TokenTypeERC20  TokenType = "ERC-20"
	TokenTypeBEP20  TokenType = "BEP-20"
)

type BlockchainNetwork string

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "DEPOSIT"
	TransactionTypeWithdraw TransactionType = "WITHDRAW"
	TransactionTypeTransfer TransactionType = "TRANSFER"
)

type EthereumNetwork BlockchainNetwork
type TronNetwork BlockchainNetwork
type BSCNetwork BlockchainNetwork
type BitcoinNetwork BlockchainNetwork
type SolanaNetwork BlockchainNetwork
type XRPLNetwork BlockchainNetwork

const (
	TronTestnetNile        TronNetwork     = "NILE"
	EthereumTestnetSepolia EthereumNetwork = "SEPOLIA"
	BSCTestnet             BSCNetwork      = "TESTNET"
	TronMainnet            TronNetwork     = "MAINNET"
	EthereumMainnet        EthereumNetwork = "MAINNET"
	BSCMainnet             BSCNetwork      = "MAINNET"
	BitcoinSignet          BitcoinNetwork  = "SIGNET"
	BitcoinMainnet         BitcoinNetwork  = "MAINNET"
	SolanaDevnet           SolanaNetwork   = "DEVNET"
	SolanaMainnet          SolanaNetwork   = "MAINNET"
	XRPLTestnet            XRPLNetwork     = "TESTNET"
	XRPLMainnet            XRPLNetwork     = "MAINNET"
)

type TronWalletBalance struct {
	TRXBalance         decimal.Decimal `json:"trx_balance"`
	AvailableEnergy    int64           `json:"available_energy"`
	TotalEnergy        int64           `json:"total_energy"`
	AvailableBandwidth int64           `json:"available_bandwidth"`
	TotalBandwidth     int64           `json:"total_bandwidth"`
	TRC20Balances      []TRC20Balance  `json:"trc20_balances"`
	TRC10Balances      []TRC10Balance  `json:"trc10_balances"`
}

type EthereumWalletBalance struct {
	ETHBalance    decimal.Decimal `json:"eth_balance"`
	ERC20Balances []ERC20Balance  `json:"erc20_balances"`
}

type BitcoinWalletBalance struct {
	Balance decimal.Decimal `json:"btc_balance"`
}

type BSCWalletBalance struct {
	BNBBalance    decimal.Decimal `json:"bnb_balance"`
	BEP20Balances []BEP20Balance  `json:"bep20_balances"`
}

type XRPLWalletBalance struct {
	Balance string `json:"xrp_balance"`
}

type SolanaWalletBalance struct {
	SOLBalance decimal.Decimal `json:"sol_balance"`
}

type BEP20Balance struct {
	ContractAddress bsc_address.BSCAddress `json:"contract_address"`
	Balance         decimal.Decimal        `json:"balance"`
}

type ERC20Balance struct {
	ContractAddress ethereum_address.EthereumAddress `json:"contract_address"`
	Balance         decimal.Decimal                  `json:"balance"`
}

type TRC20Balance struct {
	ContractAddress tron_address.TronAddress `json:"contract_address"`
	Balance         decimal.Decimal          `json:"balance"`
}

type TRC10Balance struct {
	TokenID string          `json:"token_id"`
	Balance decimal.Decimal `json:"balance"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	Result        bool   `json:"result"`
	Code          string `json:"code"`
	Message       string `json:"message"`
}

type BSCWithdrawRequest struct {
	ContractAddress *bsc_address.BSCAddress `json:"contract_address,omitempty"`
	PrivateKey      string                  `json:"private_key"`
	ToAddress       bsc_address.BSCAddress  `json:"to_address"`
	Amount          decimal.Decimal         `json:"amount"`
}

type BSCWithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
}

type BitcoinWithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
}

//**** SQL Models ****//

type Token struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	TokenID      *string
	Type         TokenType
	Name         string
	Symbol       string
	IsActive     bool
	BlockchainID uuid.UUID `gorm:"type:uuid"`
	LogoURL      string
	YahooSymbol  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Blockchain   Blockchain `gorm:"foreignKey:BlockchainID"`
}

type Blockchain struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string
	Symbol      string
	Network     BlockchainNetwork
	IsActive    bool
	LogoURL     string
	ExplorerURL string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tokens      []Token `gorm:"foreignKey:BlockchainID"`
}

type BlockchainParse struct {
	BlockchainID    uuid.UUID  `gorm:"type:uuid;primary_key"`
	Blockchain      Blockchain `gorm:"foreignKey:BlockchainID"`
	LastBlockNumber *uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Wallet struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	Address        string
	MetaID         uuid.UUID `gorm:"type:uuid"`
	Meta           WalletMeta
	Index          uint32
	DerivationPath string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type WalletMeta struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	MainWallet   string
	LastIndex    uint32
	BlockchainID uuid.UUID  `gorm:"type:uuid"`
	Blockchain   Blockchain `gorm:"foreignKey:BlockchainID"`
}

type Transaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	TxID      string
	TokenID   uuid.UUID `gorm:"type:uuid"`
	Token     Token     `gorm:"foreignKey:TokenID"`
	ToAddress string
	Amount    decimal.Decimal
	Fee       decimal.Decimal
	Type      TransactionType
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SweepingSessionStatus string

const (
	SweepingSessionStatusPending   SweepingSessionStatus = "PENDING"
	SweepingSessionStatusRunning   SweepingSessionStatus = "RUNNING"
	SweepingSessionStatusCompleted SweepingSessionStatus = "COMPLETED"
	SweepingSessionStatusFailed    SweepingSessionStatus = "FAILED"
	SweepingSessionStatusCancelled SweepingSessionStatus = "CANCELLED"
)

type SweepingSession struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	WalletMetaID       uuid.UUID `gorm:"type:uuid"`
	TokenID            uuid.UUID `gorm:"type:uuid"`
	MinAmountThreshold decimal.Decimal
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Status             SweepingSessionStatus
	Meta               *datatypes.JSON `gorm:"type:jsonb"`
}

type Sweep struct {
	TransactionID     uuid.UUID `gorm:"type:uuid"`
	SweepingSessionID uuid.UUID `gorm:"type:uuid"`
	ErrorMessage      *string
	CreatedAt         time.Time
}

type UTXO struct {
	ID                uuid.UUID   `gorm:"type:uuid;primary_key"`
	TransactionID     uuid.UUID   `gorm:"type:uuid"`
	Transaction       Transaction `gorm:"foreignKey:TransactionID"`
	Address           string
	Vout              int
	ScriptPubKeyBytes []byte `gorm:"column:scriptpubkeybytes"`
	Amount            decimal.Decimal
}

func (Blockchain) TableName() string {
	return "blockchains"
}
