package solana

import (
	"context"
	"fiatless/pkg/solana/address"
	"fmt"
	"log"
	"math/big"

	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/constants"

	bin "github.com/gagliardetto/binary"

	solana_go "github.com/gagliardetto/solana-go"
	solana_system "github.com/gagliardetto/solana-go/programs/system"
	solana_rpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/shopspring/decimal"
)

type Solana struct {
	rpcClient *solana_rpc.Client
}

func NewSolana(rpcClient *solana_rpc.Client) *Solana {
	return &Solana{rpcClient: rpcClient}
}

func (c *Solana) RPC() *solana_rpc.Client { return c.rpcClient }

func (c *Solana) GetLatestSlot(ctx context.Context) (uint64, error) {
	slot, err := c.rpcClient.GetSlot(ctx, solana_rpc.CommitmentFinalized)
	if err != nil {
		return 0, err
	}
	return slot, nil
}

func (c *Solana) GetBlocks(ctx context.Context, start uint64, end uint64) ([]*solana_rpc.GetBlockResult, error) {
	results := []*solana_rpc.GetBlockResult{}
	var v0 uint64 = 0
	for slot := start; slot <= end; slot++ {
		blk, err := c.rpcClient.GetBlockWithOpts(ctx, slot, &solana_rpc.GetBlockOpts{
			Encoding:                       solana_go.EncodingBase64,
			Commitment:                     solana_rpc.CommitmentFinalized,
			TransactionDetails:             solana_rpc.TransactionDetailsFull,
			MaxSupportedTransactionVersion: &v0,
		})
		log.Println("slot", slot, "blk", blk, "err", err)
		if err != nil || blk == nil {
			continue
		}
		results = append(results, blk)
	}
	return results, nil
}

func (c *Solana) ParseBlock(ctx context.Context, walletAddresses []address.SolanaAddress, tokenID string, block *solana_rpc.GetBlockResult) ([]bp_models.Deposit, error) {
	if block == nil {
		return nil, nil
	}
	addrSet := map[string]struct{}{}
	for _, a := range walletAddresses {
		addrSet[a.String()] = struct{}{}
	}
	deposits := []bp_models.Deposit{}
	for _, etx := range block.Transactions {
		if etx.Meta == nil || etx.Transaction == nil {
			continue
		}
		binTx := etx.Transaction.GetBinary()
		if binTx == nil {
			continue
		}
		tx, err := solana_go.TransactionFromDecoder(bin.NewBinDecoder(binTx))
		if err != nil {
			continue
		}
		for i := range etx.Meta.PostBalances {
			if etx.Meta.PostBalances[i] > etx.Meta.PreBalances[i] {
				if i >= len(tx.Message.AccountKeys) {
					continue
				}
				addrStr := tx.Message.AccountKeys[i].String()
				if _, ok := addrSet[addrStr]; !ok {
					continue
				}
				inc := etx.Meta.PostBalances[i] - etx.Meta.PreBalances[i]
				amount := decimal.NewFromBigInt(new(big.Int).SetUint64(inc), -int32(constants.SolanaDecimals)).String()
				ts := int64(0)
				if block.BlockTime != nil {
					ts = int64(*block.BlockTime)
				}
				deposits = append(deposits, bp_models.Deposit{
					TxID:      tx.Signatures[0].String(),
					TokenID:   tokenID,
					ToAddress: addrStr,
					Amount:    amount,
					Timestamp: ts,
				})
			}
		}
	}
	return deposits, nil
}

func (c *Solana) GetBalance(ctx context.Context, addr address.SolanaAddress) (uint64, error) {
	pub := solana_go.MustPublicKeyFromBase58(addr.String())
	out, err := c.rpcClient.GetBalance(ctx, pub, solana_rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("getBalance: %w", err)
	}
	return uint64(out.Value), nil
}

func (c *Solana) LamportsToSOL(lamports uint64) decimal.Decimal {
	return decimal.NewFromInt(int64(lamports)).Shift(-int32(constants.SolanaDecimals))
}

func (c *Solana) TransferSOL(ctx context.Context, privKey string, to address.SolanaAddress, amount decimal.Decimal) (map[string]string, error) {
	priv, err := solana_go.PrivateKeyFromBase58(privKey)
	if err != nil {
		return nil, fmt.Errorf("private key from base58: %w", err)
	}
	payer := priv.PublicKey()
	log.Println("payer", payer)
	recent, err := c.rpcClient.GetLatestBlockhash(ctx, solana_rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("getLatestBlockhash: %w", err)
	}

	toPub := solana_go.MustPublicKeyFromBase58(to.String())
	lamports := amount.Shift(int32(constants.SolanaDecimals)).BigInt().Uint64()
	ix := solana_system.NewTransferInstruction(lamports, payer, toPub).Build()

	tx, err := solana_go.NewTransaction(
		[]solana_go.Instruction{ix},
		recent.Value.Blockhash,
		solana_go.TransactionPayer(payer),
	)
	if err != nil {
		return nil, fmt.Errorf("new transaction: %w", err)
	}

	_, err = tx.Sign(func(key solana_go.PublicKey) *solana_go.PrivateKey {
		if key.Equals(payer) {
			pk := solana_go.PrivateKey(priv)
			return &pk
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	sig, err := c.rpcClient.SendTransactionWithOpts(ctx, tx, solana_rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: solana_rpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, fmt.Errorf("sendTransaction: %w", err)
	}
	return map[string]string{"transaction_id": sig.String()}, nil
}

func (c *Solana) GetWalletSignatures(ctx context.Context, walletAddress address.SolanaAddress, latestTransactionId *string) ([]solana_go.Signature, error) {
	until := solana_go.Signature{}
	if latestTransactionId != nil {
		until, _ = solana_go.SignatureFromBase58(*latestTransactionId)
	}
	limit := constants.SolanaLimitTransactions
	signatures, err := c.rpcClient.GetSignaturesForAddressWithOpts(ctx, solana_go.PublicKeyFromBytes(walletAddress.Raw()), &solana_rpc.GetSignaturesForAddressOpts{
		Limit: &limit,
		Until: until,
	})
	if err != nil {
		return nil, fmt.Errorf("getSignaturesForAddress: %w", err)
	}
	signaturesStr := []solana_go.Signature{}
	for _, signature := range signatures {
		signaturesStr = append(signaturesStr, signature.Signature)
	}
	return signaturesStr, nil
}

func (c *Solana) GetTransaction(ctx context.Context, signature solana_go.Signature) (*solana_rpc.GetTransactionResult, error) {
	transaction, err := c.rpcClient.GetTransaction(ctx, signature, &solana_rpc.GetTransactionOpts{
		Commitment: solana_rpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, fmt.Errorf("getTransaction: %w", err)
	}
	return transaction, nil
}

func (c *Solana) GetTransactions(ctx context.Context, signatures []solana_go.Signature) ([]*solana_rpc.GetTransactionResult, error) {
	transactions := []*solana_rpc.GetTransactionResult{}
	for _, signature := range signatures {
		transaction, err := c.GetTransaction(ctx, signature)
		if err != nil {
			return nil, fmt.Errorf("getTransaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
