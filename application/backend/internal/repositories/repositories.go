package repositories

import (
	"gorm.io/gorm"
)

type Repositories struct {
	Token           TokenRepository
	Blockchain      BlockchainRepository
	Wallet          WalletRepository
	WalletMeta      WalletMetaRepository
	Transaction     TransactionRepository
	BlockchainParse BlockchainParseRepository
	SweepingSession SweepingSessionRepository
	Sweep           SweepRepository
	UTXO            UTXORepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Token:           NewSQLTokenRepository(db),
		Blockchain:      NewSQLBlockchainRepository(db),
		Wallet:          NewSQLWalletRepository(db),
		WalletMeta:      NewSQLWalletMetaRepository(db),
		Transaction:     NewSQLTransactionRepository(db),
		BlockchainParse: NewSQLBlockchainParseRepository(db),
		SweepingSession: NewSQLSweepingSessionRepository(db),
		Sweep:           NewSQLSweepRepository(db),
		UTXO:            NewSQLUTXORepository(db),
	}
}
