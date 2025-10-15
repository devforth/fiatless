package tron

import (
	"context"
	"errors"
	"fiatless/internal/models"
	"fiatless/internal/repositories"
	"fiatless/pkg/tron/address"
	"fiatless/pkg/tron/wallet"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SweepingTask struct {
	sweepingSession *models.SweepingSession
	wallets         []*wallet.TronWallet
	mainWallet      wallet.TronWallet
	trxHolderWallet wallet.TronWallet
	repositories    *repositories.Repositories
	trxTokenID      uuid.UUID
}

func NewSweepingTask(sweepingSession *models.SweepingSession, wallets []*wallet.TronWallet, mainWallet wallet.TronWallet, trxHolderWallet wallet.TronWallet, repositories *repositories.Repositories, trxTokenID uuid.UUID) *SweepingTask {
	return &SweepingTask{
		sweepingSession: sweepingSession,
		wallets:         wallets,
		mainWallet:      mainWallet,
		trxHolderWallet: trxHolderWallet,
		repositories:    repositories,
		trxTokenID:      trxTokenID,
	}
}

func (t *SweepingTask) Do(ctx context.Context) error {
	tokenID := t.sweepingSession.TokenID

	token, err := t.repositories.Token.GetToken(ctx, &models.Token{
		ID: tokenID,
	})
	if err != nil {
		return err
	}
	tokenAddress, err := address.NewTronAddressFromBase58(*token.TokenID)
	if err != nil {
		return err
	}

	trxHolderWallet := t.trxHolderWallet

	for _, wallet := range t.wallets {
		tokenBalance, err := wallet.GetBalanceOffChain(t.repositories.Wallet, token.ID)
		if err != nil {
			return err
		}
		log.Println("Token balance", tokenBalance)
		if tokenBalance.LessThan(t.sweepingSession.MinAmountThreshold) {
			continue
		}

		trxAmount, err := trxHolderWallet.GetBalanceOffChain(t.repositories.Wallet, t.trxTokenID)
		if err != nil {
			return err
		}

		if trxAmount.LessThan(decimal.NewFromInt(0)) {
			continue
		}

		if trxHolderWallet.GetAddress().Equals(t.mainWallet.GetAddress()) {
			// send 50% of the TRX balance to the wallet for sweeping
			amount := trxAmount.Div(decimal.NewFromInt(2))
			trxAmount = &amount
		}
		withdrawal, err := trxHolderWallet.Withdraw(wallet.GetAddress(), *trxAmount, nil)

		var errorMessage *string
		if err != nil {
			errStr := err.Error()
			errorMessage = &errStr
		}

		tokenTransactionId := uuid.New()
		if errorMessage != nil {
			// t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
			// 	ID:        uuid.New(),
			// 	TxID:      withdrawal.TransactionID,
			// 	Fee:       withdrawal.Fee,
			// 	TokenID:   token.ID,
			// 	ToAddress: wallet.GetAddress().String(),
			// 	Amount:    *trxAmount,
			// 	Type:      models.TransactionTypeWithdraw,
			// 	CreatedAt: time.Now(),
			// 	UpdatedAt: time.Now(),
			// })
			withdrawal, err = wallet.Withdraw(t.mainWallet.GetAddress(), *tokenBalance, tokenAddress)
			if err != nil {
				errStr := err.Error()
				errorMessage = &errStr
			}

			t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
				ID:        tokenTransactionId,
				TxID:      withdrawal.TransactionID,
				Fee:       withdrawal.Fee,
				TokenID:   token.ID,
				ToAddress: trxHolderWallet.GetAddress().String(),
				Amount:    *tokenBalance,
				Type:      models.TransactionTypeWithdraw,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}

		t.repositories.Sweep.CreateSweep(ctx, &models.Sweep{
			TransactionID:     tokenTransactionId,
			SweepingSessionID: t.sweepingSession.ID,
			ErrorMessage:      errorMessage,
			CreatedAt:         time.Now(),
		})

		if errorMessage != nil {
			return err
		}
	}

	// 4. Withdraw TRX to main wallet
	trxBalance, err := trxHolderWallet.GetBalanceOffChain(t.repositories.Wallet, t.trxTokenID)
	if err != nil {
		return err
	}

	if trxBalance.LessThan(decimal.NewFromInt(0)) {
		return errors.New("TRX balance on TRX holder wallet is 0")
	}

	if !trxHolderWallet.GetAddress().Equals(t.mainWallet.GetAddress()) {
		withdrawal, err := trxHolderWallet.Withdraw(t.mainWallet.GetAddress(), *trxBalance, nil)
		if err != nil {
			return err
		}

		t.repositories.Transaction.CreateTransaction(ctx, &models.Transaction{
			ID:        uuid.New(),
			TxID:      withdrawal.TransactionID,
			Fee:       withdrawal.Fee,
			TokenID:   t.trxTokenID,
			ToAddress: t.mainWallet.GetAddress().String(),
			Amount:    *trxBalance,
			Type:      models.TransactionTypeWithdraw,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// 5. Update sweeping session status to completed
	err = t.repositories.SweepingSession.UpdateSweepingSessionStatus(ctx, t.sweepingSession.ID, models.SweepingSessionStatusCompleted)
	if err != nil {
		return err
	}
	return nil
}
