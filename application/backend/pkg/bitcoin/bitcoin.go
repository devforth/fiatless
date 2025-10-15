package bitcoin

import (
	"bytes"
	"context"
	"encoding/hex"
	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/bitcoin/address"
	"fiatless/pkg/bitcoin/client"
	"log"
	"strconv"

	"github.com/shopspring/decimal"
)

type Bitcoin struct {
	Client *client.BitcoinClient
}

func NewBitcoin(client *client.BitcoinClient) *Bitcoin {
	return &Bitcoin{Client: client}
}

func (b *Bitcoin) GetBalance(address address.BitcoinAddress) (decimal.Decimal, error) {
	return b.Client.GetBalance(address.String())
}

// GetBlocks retrieves blocks within a specified height range and returns them as BitcoinBlock structs
// startBlockNumber: starting block height (inclusive)
// endBlockNumber: ending block height (inclusive)
func (b *Bitcoin) GetBlocks(startBlockNumber, endBlockNumber uint64, verbose int) ([]*BitcoinBlock, error) {
	// Use verbose level 2 to get full transaction details
	blocksData, err := b.Client.GetBlocks(startBlockNumber, endBlockNumber, verbose)
	if err != nil {
		return nil, err
	}

	var blocks []*BitcoinBlock
	for _, blockData := range blocksData {
		block, err := b.parseBlockData(blockData)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (b *Bitcoin) GetBlockCount() (uint64, error) {
	return b.Client.GetBlockCount()
}

// parseBlockData converts raw block data from RPC to BitcoinBlock struct
func (b *Bitcoin) parseBlockData(blockData map[string]interface{}) (*BitcoinBlock, error) {
	// Unwrap nested result if present (some callers may pass the full RPC envelope)
	if result, ok := blockData["result"]; ok {
		if resultMap, ok := result.(map[string]interface{}); ok {
			blockData = resultMap
		}
	}
	block := &BitcoinBlock{}
	log.Println("blockData", blockData)
	// Parse hash
	if hash, ok := blockData["hash"].(string); ok {
		block.Hash = hash
	}

	// Parse height
	if heightFloat, ok := blockData["height"].(uint64); ok {
		block.Height = int64(heightFloat)
	} else {
		log.Printf("height err: expected uint64, got %T with value %v", blockData["height"], blockData["height"])
	}

	// Parse time
	if timeFloat, ok := blockData["time"].(float64); ok {
		block.Time = int64(timeFloat)
	}

	// Parse transactions
	if txData, ok := blockData["tx"].([]interface{}); ok {
		for _, tx := range txData {
			if txMap, ok := tx.(map[string]interface{}); ok {
				transaction, err := b.parseTransactionData(txMap)
				if err != nil {
					return nil, err
				}
				block.Transactions = append(block.Transactions, *transaction)
			}
		}
	}

	return block, nil
}

// Withdraw creates, signs and returns a native BTC withdrawal transaction id and fee.
// It uses the provided private key, target address and amount. UTXOs are discovered
// by querying listunspent for the derived address from the key. For production, wire
// in DB UTXOs to have exact control (recommended).
func (b *Bitcoin) Withdraw(to *address.BitcoinAddress, amount decimal.Decimal, utxos []bp_models.UTXO) (bp_models.BitcoinWithdrawResponse, error) {
	// Build per-input coins (with private keys)
	coins, err := convertBpUtxosToCoins(utxos)
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}

	// Build withdraw processor; network/address type selection should align with your config
	proc := NewWithdrawProcessor(nil, models.BitcoinSignet, address.P2TR)

	// Build and sign transaction using per-input keys
	tx, _, err := proc.CreateAndSignWithPerInputKeys(context.Background(), coins, *to, amount)
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}

	// Serialize and broadcast via sendrawtransaction
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}

	// If you have an RPC that accepts hex, convert here (placeholder call - implement on client)
	hex := hex.EncodeToString(buf.Bytes())
	log.Println("hex", hex)
	txid, err := b.Client.SendRawTransaction(hex) // For now we only return the txid computed locally
	if err != nil {
		return bp_models.BitcoinWithdrawResponse{}, err
	}

	return bp_models.BitcoinWithdrawResponse{TransactionID: txid}, nil
}

// parseTransactionData converts raw transaction data to BitcoinTransaction struct
func (b *Bitcoin) parseTransactionData(txData map[string]interface{}) (*BitcoinTransaction, error) {
	// Unwrap nested result if present
	if result, ok := txData["result"]; ok {
		if resultMap, ok := result.(map[string]interface{}); ok {
			txData = resultMap
		}
	}
	transaction := &BitcoinTransaction{}

	// Parse txid
	if txid, ok := txData["txid"].(string); ok {
		transaction.TxID = txid
	}

	// Parse fee
	if fee, ok := txData["fee"].(float64); ok {
		transaction.Fee = &fee
	}

	// Parse vin (inputs)
	if vinData, ok := txData["vin"].([]interface{}); ok {
		var inputs []BitcoinTransactionInput
		for _, vin := range vinData {
			if vinMap, ok := vin.(map[string]interface{}); ok {
				input := BitcoinTransactionInput{}
				if txid, ok := vinMap["txid"].(string); ok {
					input.TxID = &txid
				}
				if vout, ok := vinMap["vout"].(float64); ok {
					voutInt := int(vout)
					input.Vout = &voutInt
				}
				inputs = append(inputs, input)
			}
		}
		if len(inputs) > 0 {
			transaction.Vin = &inputs
		}
	}

	// Parse vout (outputs)
	if voutData, ok := txData["vout"].([]interface{}); ok {
		var outputs []BitcoinTransactionOutput
		for _, vout := range voutData {
			if voutMap, ok := vout.(map[string]interface{}); ok {
				output := BitcoinTransactionOutput{}

				if value, ok := voutMap["value"].(float64); ok {
					output.Value = &value
				}

				if n, ok := voutMap["n"].(float64); ok {
					nInt := int(n)
					output.N = &nInt
				}

				if scriptPubKey, ok := voutMap["scriptPubKey"].(map[string]interface{}); ok {
					spk := &BitcoinTransactionScriptPubKey{}
					if address, ok := scriptPubKey["address"].(string); ok {
						spk.Address = &address
					}
					if hex, ok := scriptPubKey["hex"].(string); ok {
						spk.Hex = &hex
					}
					output.ScriptPubKey = spk
				}

				outputs = append(outputs, output)
			}
		}
		if len(outputs) > 0 {
			transaction.Vout = &outputs
		}
	}

	return transaction, nil
}

// return utxo that was spent and deposits
// txIDwithvout - txid-vout
func (b *Bitcoin) ParseBlock(wallets []address.BitcoinAddress, txIDwithvout map[string]struct{}, block *BitcoinBlock) ([]bp_models.BitcoinTransaction, error) {

	transactions := []bp_models.BitcoinTransaction{}
	// Create a map of wallet addresses for O(1) lookup and deduplication
	walletMap := make(map[string]bool)
	for _, wallet := range wallets {
		walletMap[wallet.String()] = true
	}

	for _, transaction := range block.Transactions {
		deposits := []bp_models.ReceivedUTXO{}
		spentUTXOs := []bp_models.SpentUTXO{}
		if transaction.Vin != nil {
			for _, input := range *transaction.Vin {
				if input.TxID != nil && input.Vout != nil {
					if _, exists := txIDwithvout[*input.TxID+"-"+strconv.Itoa(*input.Vout)]; exists {
						spentUTXOs = append(spentUTXOs, bp_models.SpentUTXO{
							TxID: *input.TxID,
							Vout: *input.Vout,
						})
					}
				}
			}
		}

		if transaction.Vout != nil {
			for _, output := range *transaction.Vout {
				if output.ScriptPubKey != nil && output.ScriptPubKey.Address != nil && output.Value != nil && output.N != nil && output.ScriptPubKey.Hex != nil {
					// Check if this address is one of our wallets using the map
					if walletMap[*output.ScriptPubKey.Address] {
						deposits = append(deposits, bp_models.ReceivedUTXO{
							TxID:            transaction.TxID,
							Amount:          decimal.NewFromFloat(*output.Value).String(),
							Vout:            *output.N,
							Scriptpubkeyhex: *output.ScriptPubKey.Hex,
							Address:         *output.ScriptPubKey.Address,
						})
					}
				}
			}
		}

		if len(deposits) > 0 || len(spentUTXOs) > 0 {
			fee := 0.0
			if transaction.Fee != nil {
				fee = *transaction.Fee
			}
			transactions = append(transactions, bp_models.BitcoinTransaction{
				TxID: transaction.TxID,
				Vin:  spentUTXOs,
				Vout: deposits,
				Fee:  fee,
				Time: int64(block.Time),
			})
		}
	}
	return transactions, nil
}
