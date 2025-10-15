package client

import (
	"bytes"
	"encoding/json"
	"fiatless/pkg/httpclient"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/shopspring/decimal"
)

type BitcoinRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int64       `json:"id"`
}

type BitcoinRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  json.RawMessage  `json:"result"`
	Error   *BitcoinRPCError `json:"error"`
	ID      int64            `json:"id"`
}

type BitcoinRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BitcoinRPCError) Error() string {
	return fmt.Sprintf("Bitcoin RPC error %d: %s", e.Code, e.Message)
}

type BitcoinClient struct {
	httpClient    *http.Client
	rateTransport *httpclient.RateLimitedTransport
	endpoint      string
	requestID     int64
}

func NewBitcoinClient(endpoint string, rps float64) (*BitcoinClient, error) {
	rateTransport := httpclient.NewRateLimitedTransport(rps, nil)
	httpClient := &http.Client{
		Transport: rateTransport,
	}

	return &BitcoinClient{
		httpClient:    httpClient,
		rateTransport: rateTransport,
		endpoint:      endpoint,
	}, nil
}

func (c *BitcoinClient) UpdateRateLimit(rps float64) {
	c.rateTransport.UpdateRPS(rps)
}

func (c *BitcoinClient) Call(method string, params interface{}) (*BitcoinRPCResponse, error) {
	requestID := atomic.AddInt64(&c.requestID, 1)

	request := BitcoinRPCRequest{
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
		ID:      requestID,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response BitcoinRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	return &response, nil
}

func (c *BitcoinClient) GetBalance(address string) (decimal.Decimal, error) {
	params := []interface{}{0, 9999999, []string{address}}

	response, err := c.Call("listunspent", params)
	if err != nil {
		return decimal.Zero, fmt.Errorf("listunspent failed: %v", err)
	}

	var utxos []struct {
		Amount decimal.Decimal `json:"amount"`
	}

	if err := json.Unmarshal(response.Result, &utxos); err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse listunspent response: %v", err)
	}

	balance := decimal.Zero
	for _, utxo := range utxos {
		balance = balance.Add(utxo.Amount)
	}

	return balance, nil
}

func (c *BitcoinClient) GetWalletBalance() (decimal.Decimal, error) {
	response, err := c.Call("getbalance", []interface{}{})
	if err != nil {
		return decimal.Zero, fmt.Errorf("getbalance failed: %v", err)
	}

	var balance decimal.Decimal
	if err := json.Unmarshal(response.Result, &balance); err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse balance: %v", err)
	}

	return balance, nil
}

func (c *BitcoinClient) GetBlockchainInfo() (map[string]interface{}, error) {
	response, err := c.Call("getblockchaininfo", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("getblockchaininfo failed: %v", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(response.Result, &info); err != nil {
		return nil, fmt.Errorf("failed to parse blockchain info: %v", err)
	}

	return info, nil
}

func (c *BitcoinClient) GetBlock(blockHash string, verbose int) (map[string]interface{}, error) {
	response, err := c.Call("getblock", []interface{}{blockHash, verbose})
	if err != nil {
		return nil, fmt.Errorf("getblock failed: %v", err)
	}

	var block map[string]interface{}
	if err := json.Unmarshal(response.Result, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %v", err)
	}

	return block, nil
}

func (c *BitcoinClient) GetBlockCount() (uint64, error) {
	response, err := c.Call("getblockcount", []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("getblockcount failed: %v", err)
	}

	var count uint64
	if err := json.Unmarshal(response.Result, &count); err != nil {
		return 0, fmt.Errorf("failed to parse block count: %v", err)
	}

	return count, nil
}

func (c *BitcoinClient) GetBlockHash(blockHeight int) (string, error) {
	response, err := c.Call("getblockhash", []interface{}{blockHeight})
	if err != nil {
		return "", fmt.Errorf("getblockhash failed: %v", err)
	}

	var hash string
	if err := json.Unmarshal(response.Result, &hash); err != nil {
		return "", fmt.Errorf("failed to parse block hash: %v", err)
	}

	return hash, nil
}

// SendRawTransaction broadcasts a raw transaction (hex-encoded) to the network
// and returns its txid on success.
func (c *BitcoinClient) SendRawTransaction(txHex string) (string, error) {
	response, err := c.Call("sendrawtransaction", []interface{}{txHex})
	if err != nil {
		return "", fmt.Errorf("sendrawtransaction failed: %v", err)
	}

	var txid string
	if err := json.Unmarshal(response.Result, &txid); err != nil {
		return "", fmt.Errorf("failed to parse sendrawtransaction response: %v", err)
	}

	return txid, nil
}

// GetBlocks retrieves blocks within a specified height range
// minHeight: minimum block height (inclusive)
// maxHeight: maximum block height (inclusive)
// verbose: verbosity level for block data (0=hex, 1=json, 2=json with tx details)
func (c *BitcoinClient) GetBlocks(minHeight, maxHeight uint64, verbose int) ([]map[string]interface{}, error) {
	if minHeight > maxHeight {
		return nil, fmt.Errorf("minHeight (%d) cannot be greater than maxHeight (%d)", minHeight, maxHeight)
	}

	// Limit the range to prevent excessive requests
	const maxBlocksPerRequest = 1000
	if maxHeight-minHeight+1 > maxBlocksPerRequest {
		return nil, fmt.Errorf("block range too large, maximum %d blocks allowed per request", maxBlocksPerRequest)
	}

	var blocks []map[string]interface{}

	for height := minHeight; height <= maxHeight; height++ {
		// Get block hash for this height
		blockHash, err := c.GetBlockHash(int(height))
		if err != nil {
			return nil, fmt.Errorf("failed to get block hash for height %d: %v", height, err)
		}

		// Get block data
		block, err := c.GetBlock(blockHash, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d (hash: %s): %v", height, blockHash, err)
		}

		// Ensure height is included in the block data
		block["height"] = height
		blocks = append(blocks, block)
	}

	return blocks, nil
}
