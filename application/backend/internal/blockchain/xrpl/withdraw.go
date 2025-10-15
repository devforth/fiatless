package xrpl

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/constants"
	"fiatless/internal/ijson"
	"fiatless/pkg/xrpl/wallet"
	"fmt"
	"time"

	xrpl_hash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	xrpl_rpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	xrpl_types "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/shopspring/decimal"
)

type WithdrawHandler struct {
	handler.XRPLHandler
}

func NewWithdrawHandler(xrpl *xrpl_rpc.Client) *WithdrawHandler {
	return &WithdrawHandler{
		XRPLHandler: handler.NewXRPLHandler(xrpl),
	}
}

func (h *WithdrawHandler) CommandPath() string {
	return "/xrpl/wallet/withdraw"
}

func (h *WithdrawHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.XRPLWithdrawRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	wallet, err := wallet.NewBaseXRPLWalletFromPrivateKey(params.PrivateKey)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	tx, err := h.withdraw(wallet, params.ToAddress, params.Amount)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	return client.SendResult(map[string]any{
		"id":     requestID,
		"result": tx,
	})
}

func (h *WithdrawHandler) withdraw(wallet *wallet.BaseXRPLWallet, toAddress xrpl_types.Address, amount decimal.Decimal) (bp_models.XRPLWithdrawResponse, error) {
	xrpAmountInt := amount.Shift(int32(constants.XRPLDecimals)).BigInt().Uint64()
	fee, err := h.XRPL.GetFee(&server.FeeRequest{})
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to get fee: %v", err)
	}
	p := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: wallet.GetAddress(),
			Fee:     fee.Drops.OpenLedgerFee,
		},
		Destination: toAddress,
		Amount:      xrpl_types.XRPCurrencyAmount(xrpAmountInt),
		DeliverMax:  xrpl_types.XRPCurrencyAmount(xrpAmountInt),
	}

	flattenedTx := p.Flatten()

	if err := h.XRPL.Autofill(&flattenedTx); err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to autofill: %v", err)
	}

	txBlob, err := wallet.SignTransaction(flattenedTx)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Submit the blob (do not use built-in wait; it may time out too aggressively)
	submitRes, err := h.XRPL.SubmitTxBlob(txBlob, false)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("submit failed: %v", err)
	}
	if submitRes.EngineResult != "tesSUCCESS" && submitRes.EngineResult != "terQUEUED" {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("submit engine result: %s", submitRes.EngineResult)
	}

	// Compute transaction hash
	txHash, err := xrpl_hash.SignTxBlob(txBlob)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, fmt.Errorf("failed to compute tx hash: %v", err)
	}

	// Extract LastLedgerSequence from the prepared tx if present
	var lastLedgerSeq uint32
	if v, ok := flattenedTx["LastLedgerSequence"].(uint32); ok {
		lastLedgerSeq = v
	}

	// Wait until validated or expired
	res, err := h.waitTxValidated(txHash, lastLedgerSeq, 60, time.Second)
	if err != nil {
		return bp_models.XRPLWithdrawResponse{}, err
	}

	return bp_models.XRPLWithdrawResponse{
		TransactionID: res.Hash.String(),
	}, nil
}

// waitTxValidated polls tx by hash until it is validated or LastLedgerSequence is reached (if provided).
func (h *WithdrawHandler) waitTxValidated(txHash string, lastLedgerSequence uint32, maxTries int, delay time.Duration) (*requests.TxResponse, error) {
	tries := 0
	for tries < maxTries {
		// Query transaction by hash
		raw, err := h.XRPL.Request(&requests.TxRequest{Transaction: txHash})
		if err == nil && raw != nil {
			var r requests.TxResponse
			if err := raw.GetResult(&r); err == nil {
				if r.Validated {
					return &r, nil
				}
			}
		}

		// If we have a last ledger sequence and it's passed, stop
		if lastLedgerSequence != 0 {
			idx, err := h.XRPL.GetLedgerIndex()
			if err == nil && idx.Uint32() >= lastLedgerSequence {
				break
			}
		}

		time.Sleep(delay)
		tries++
	}
	return nil, fmt.Errorf("transaction not found or not validated before LastLedgerSequence")
}
