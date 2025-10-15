package walletmgr

import (
	"context"
	"errors"
	"fiatless/internal/models"
	"fiatless/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnsureInit ensures that wallet meta and the main wallet record exist for a given blockchain and main wallet address.
// It returns the walletsMetaId.
func EnsureInit(ctx context.Context, repositories *repositories.Repositories, blockchainId uuid.UUID, mainWalletAddress string, mainWalletDerivationPath string) (uuid.UUID, error) {
	walletMeta, err := repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(ctx, blockchainId, mainWalletAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			generatedId := uuid.New()
			err = repositories.WalletMeta.CreateWalletMeta(ctx, &models.WalletMeta{
				ID:           generatedId,
				MainWallet:   mainWalletAddress,
				BlockchainID: blockchainId,
			})
			if err != nil {
				return uuid.Nil, err
			}
			walletMeta = &models.WalletMeta{ID: generatedId}
		} else {
			return uuid.Nil, err
		}
	}

	// Ensure main wallet exists
	_, err = repositories.Wallet.GetWalletByAddress(ctx, mainWalletAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = repositories.Wallet.CreateWallet(ctx, &models.Wallet{
				ID:             uuid.New(),
				Address:        mainWalletAddress,
				Index:          0,
				DerivationPath: mainWalletDerivationPath,
				MetaID:         walletMeta.ID,
			})
			if err != nil {
				return uuid.Nil, err
			}
		} else {
			return uuid.Nil, err
		}
	}

	return walletMeta.ID, nil
}

// GetLastAddressIndex loads or creates wallet meta and returns the last used address index.
func GetLastAddressIndex(ctx context.Context, repositories *repositories.Repositories, blockchainId uuid.UUID, mainWalletAddress string) (uint32, error) {
	walletMeta, err := repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(ctx, blockchainId, mainWalletAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = repositories.WalletMeta.CreateWalletMeta(ctx, &models.WalletMeta{
				ID:           uuid.New(),
				MainWallet:   mainWalletAddress,
				BlockchainID: blockchainId,
			})
			if err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, err
	}
	return walletMeta.LastIndex, nil
}

// SetLastAddressIndex updates the last address index in wallet meta.
func SetLastAddressIndex(ctx context.Context, repositories *repositories.Repositories, blockchainId uuid.UUID, mainWalletAddress string, index uint32) error {
	walletMeta, err := repositories.WalletMeta.GetWalletMetaByBlockchainIDAndWalletAddress(ctx, blockchainId, mainWalletAddress)
	if err != nil {
		return err
	}
	walletMeta.LastIndex = index
	return repositories.WalletMeta.UpdateWalletMeta(ctx, walletMeta)
}
