package bitcoin

import (
	"context"
	"fiatless/internal/blockchain/handler"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/ijson"
	"fiatless/pkg/bitcoin"
	"strconv"
)

type BlockParseHandler struct {
	handler.BitcoinHandler
}

func NewBlockParseHandler(client *bitcoin.Bitcoin) *BlockParseHandler {
	return &BlockParseHandler{
		BitcoinHandler: handler.NewBitcoinHandler(client),
	}
}

func (h *BlockParseHandler) CommandPath() string {
	return "/bitcoin/block/parse"
}

func (h *BlockParseHandler) Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error {
	var params bp_models.BitcoinParseBlocksRequest
	requestID, err := h.ParseParams(command, &params)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}

	latestBlockNumber, err := h.Bitcoin.GetBlockCount()
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	var startBlockNumber uint64 = latestBlockNumber
	if params.LatestBlockNumber != nil {
		if *params.LatestBlockNumber == latestBlockNumber {
			return h.SendErrorResult(client, requestID, "Latest block number is the same as the latest block number")
		}
		startBlockNumber = *params.LatestBlockNumber + 1
	}
	var endBlockNumber uint64 = latestBlockNumber
	if endBlockNumber-startBlockNumber > 10 {
		endBlockNumber = startBlockNumber + 10
	}
	blocks, err := h.Bitcoin.GetBlocks(startBlockNumber, endBlockNumber, 2)
	if err != nil {
		return h.SendErrorResult(client, requestID, err.Error())
	}
	wallets := params.WalletAddresses
	results := []bp_models.BitcoinTransaction{}
	parsedBlocksNumbers := []uint64{}
	failedBlocksNumbers := []uint64{}

	utxoSet := make(map[string]struct{}, len(params.UTXOs))
	for k, v := range params.UTXOs {
		utxoSet[k] = v
	}

	for _, block := range blocks {
		txs, err := h.Bitcoin.ParseBlock(wallets, utxoSet, block)
		if err != nil {
			failedBlocksNumbers = append(failedBlocksNumbers, uint64(block.Height))
			continue
		}
		results = append(results, txs...)

		// Update the rolling UTXO set with results from this block:
		// - Remove spent utxos that we tracked
		// - Add newly received utxos belonging to our wallets
		for _, tx := range txs {
			for _, spent := range tx.Vin {
				key := spent.TxID + "-" + strconv.Itoa(spent.Vout)
				delete(utxoSet, key)
			}
			for _, recv := range tx.Vout {
				key := recv.TxID + "-" + strconv.Itoa(recv.Vout)
				utxoSet[key] = struct{}{}
			}
		}
		parsedBlocksNumbers = append(parsedBlocksNumbers, uint64(block.Height))
	}

	return client.SendResult(map[string]any{
		"id": requestID,
		"result": bp_models.BitcoinParseBlocksResponse{
			ParsedBlocksNumbers: parsedBlocksNumbers,
			FailedBlocksNumbers: failedBlocksNumbers,
			Transactions:        results,
			LastBlockNumber:     endBlockNumber,
		},
	})
}
