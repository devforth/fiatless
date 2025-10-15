package tron

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/tron"
	"log"
)

type BlockParseHandler struct {
	handler.TronHandler
}

func NewBlockParseHandler(client *tron.Tron) *BlockParseHandler {
	return &BlockParseHandler{
		TronHandler: handler.NewTronHandler(client),
	}
}

func (h *BlockParseHandler) CommandPath() string {
	return "/tron/block/parse"
}

func (h *BlockParseHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.ParseBlocksRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	latestBlockNumber, err := h.Tron.GetLatestBlockNumber()
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	var startBlockNumber uint64 = latestBlockNumber
	if params.LatestBlockNumber != nil {
		startBlockNumber = *params.LatestBlockNumber + 1
	}
	var endBlockNumber uint64 = latestBlockNumber
	if endBlockNumber-startBlockNumber > 100 {
		endBlockNumber = startBlockNumber + 100
	}
	log.Println("startBlockNumber", startBlockNumber, endBlockNumber)
	blocks, err := h.Tron.GetBlocks(startBlockNumber, endBlockNumber)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	adresses := params.WalletAddresses
	results := []bp_models.Deposit{}
	parsedBlocksNumbers := []uint64{}
	failedBlocksNumbers := []uint64{}
	log.Println("tokens", params.Tokens)
	for _, block := range blocks {
		deposits, err := h.Tron.ParseBlock(ctx, adresses, params.Tokens, block)
		if err != nil {
			failedBlocksNumbers = append(failedBlocksNumbers, uint64(block.BlockHeader.RawData.Number))
			continue
		}
		results = append(results, deposits...)
		parsedBlocksNumbers = append(parsedBlocksNumbers, uint64(block.BlockHeader.RawData.Number))
	}

	return client.SendResult(map[string]any{
		"id": requestID,
		"result": bp_models.ParseBlocksResponse{
			ParsedBlocksNumbers: parsedBlocksNumbers,
			FailedBlocksNumbers: failedBlocksNumbers,
			Deposits:            results,
			LastBlockNumber:     endBlockNumber - 1,
		},
	})
}
