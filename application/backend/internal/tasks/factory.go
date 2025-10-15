package tasks

import (
	"time"

	"github.com/google/uuid"

	"fiatless/internal/models"
	"fiatless/internal/repositories"
	bitcoinService "fiatless/internal/services/bitcoin"
	solanaService "fiatless/internal/services/solana"
	tronService "fiatless/internal/services/tron"
	xrplService "fiatless/internal/services/xrpl"
	bitcoinTasks "fiatless/internal/tasks/bitcoin"
	solanaTasks "fiatless/internal/tasks/solana"
	tronTasks "fiatless/internal/tasks/tron"
	"fiatless/pkg/bitcoin"
	"fiatless/pkg/solana"
	"fiatless/pkg/tron"
	"fiatless/pkg/tron/wallet"
)

type TaskFactory struct {
	tronService    *tronService.Service
	bitcoinService *bitcoinService.Service
	solanaService  *solanaService.Service
	xrplService    *xrplService.Service
}

func NewTaskFactory(tronService *tronService.Service, bitcoinService *bitcoinService.Service, solanaService *solanaService.Service, xrplService *xrplService.Service) *TaskFactory {
	return &TaskFactory{
		tronService:    tronService,
		bitcoinService: bitcoinService,
		solanaService:  solanaService,
		xrplService:    xrplService,
	}
}

func (f *TaskFactory) CreateBitcoinBlocksTask(blockchainID uuid.UUID, timeout, cycleDelay time.Duration, walletManager *bitcoin.WalletManager, repositories *repositories.Repositories) Task {
	blocksTask := bitcoinTasks.NewBlocksTask(blockchainID, f.bitcoinService, walletManager, repositories)

	return Task{
		Name:       "BitcoinBlocks",
		Timeout:    timeout,
		CycleDelay: &cycleDelay,
		Do:         blocksTask.Do,
	}
}

func (f *TaskFactory) CreateSolanaBlocksTask(blockchainID uuid.UUID, timeout, cycleDelay time.Duration, walletManager *solana.WalletManager, repositories *repositories.Repositories) Task {
	blocksTask := solanaTasks.NewBlocksTask(blockchainID, f.solanaService, walletManager, repositories)

	return Task{
		Name:       "SolanaBlocks",
		Timeout:    timeout,
		CycleDelay: &cycleDelay,
		Do:         blocksTask.Do,
	}
}

func (f *TaskFactory) CreateSolanaTransactionsTask(blockchainID uuid.UUID, timeout, cycleDelay time.Duration, walletManager *solana.WalletManager, repositories *repositories.Repositories) Task {
	transactionsTask := solanaTasks.NewTransactionsTask(blockchainID, f.solanaService, walletManager, repositories)

	return Task{
		Name:       "SolanaTransactions",
		Timeout:    timeout,
		CycleDelay: &cycleDelay,
		Do:         transactionsTask.Do,
	}
}

func (f *TaskFactory) CreateTronBlocksTask(blockchainID uuid.UUID, timeout, cycleDelay time.Duration, walletManager *tron.WalletManager, repositories *repositories.Repositories) Task {
	blocksTask := tronTasks.NewBlocksTask(blockchainID, f.tronService, walletManager, repositories)

	return Task{
		Name:       "TronBlocks",
		Timeout:    timeout,
		CycleDelay: &cycleDelay,
		Do:         blocksTask.Do,
	}
}

func (f *TaskFactory) CreateTronSweepingTask(sweepingSession *models.SweepingSession, wallets []*wallet.TronWallet, mainWallet wallet.TronWallet, trxHolderWallet wallet.TronWallet, timeout time.Duration, repositories *repositories.Repositories, trxTokenID uuid.UUID) Task {
	sweepingTask := tronTasks.NewSweepingTask(sweepingSession, wallets, mainWallet, trxHolderWallet, repositories, trxTokenID)

	return Task{
		Name:       "TronSweeping",
		Timeout:    timeout,
		CycleDelay: nil,
		Do:         sweepingTask.Do,
	}
}
