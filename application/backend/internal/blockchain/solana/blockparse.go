package solana

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/solana"
	"log"
)

type BlockParseHandler struct {
	Base   handler.BaseHandler
	Solana *solana.Solana
}

func NewBlockParseHandler(client *solana.Solana) *BlockParseHandler {
	return &BlockParseHandler{Solana: client}
}

func (h *BlockParseHandler) CommandPath() string {
	return "/solana/block/parse"
}

func (h *BlockParseHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.SolanaParseBlocksRequest
	requestID, err := h.Base.ParseParams(command, &params)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}

	// Determine start and end slots
	latest, err := h.Solana.GetLatestSlot(ctx)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}
	var start uint64 = latest
	if params.LatestBlockNumber != nil {
		start = *params.LatestBlockNumber + 1
	}
	end := latest
	if end-start > 100 {
		end = start + 100
	}

	parsed := []uint64{}
	failed := []uint64{}
	deposits := []bp_models.Deposit{}

	addrSet := map[string]struct{}{}
	for _, a := range params.WalletAddresses {
		addrSet[a.String()] = struct{}{}
	}

	blocks, err := h.Solana.GetBlocks(ctx, start, end)
	if err != nil {
		return h.Base.SendErrorResult(client, requestID, err.Error())
	}
	for _, block := range blocks {
		deps, err := h.Solana.ParseBlock(ctx, params.WalletAddresses, params.TokenID, block)
		if err != nil {
			continue
		}
		deposits = append(deposits, deps...)
		// Append parsed slot if available; fallback to LastValidBlockHeight is not suitable; we use blockHeight.
		if block.BlockHeight != nil {
			parsed = append(parsed, uint64(*block.BlockHeight))
		}
	}

	last := end
	log.Println("solana parsed slots", parsed)
	return client.SendResult(map[string]any{
		"id": requestID,
		"result": bp_models.ParseBlocksResponse{
			ParsedBlocksNumbers: parsed,
			FailedBlocksNumbers: failed,
			Deposits:            deposits,
			LastBlockNumber:     last,
		},
	})
}
