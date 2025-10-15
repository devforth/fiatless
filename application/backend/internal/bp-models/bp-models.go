package bp_models

import (
	"fiatless/internal/models"
	bitcoin_address "fiatless/pkg/bitcoin/address"
	bsc_address "fiatless/pkg/bsc/address"
	ethereum_address "fiatless/pkg/ethereum/address"
	solana_address "fiatless/pkg/solana/address"
	tron_address "fiatless/pkg/tron/address"

	xrpl_address "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TronWalletBalanceRequest struct {
	Address          tron_address.TronAddress   `json:"address"`
	IncludeTRX       bool                       `json:"includeTRX"`
	IncludeResources bool                       `json:"includeResources"`
	TRC20Tokens      []tron_address.TronAddress `json:"trc20Tokens"`
	TRC10Tokens      []string                   `json:"trc10Tokens"`
}

type TronWithdrawRequest struct {
	ContractAddress *tron_address.TronAddress `json:"contractAddress,omitempty"`
	PrivateKey      string                    `json:"privateKey"`
	ToAddress       tron_address.TronAddress  `json:"toAddress"`
	Amount          decimal.Decimal           `json:"amount"`
	FeeLimit        int64                     `json:"feeLimit"`
}

type TronWithdrawResponse struct {
	TransactionID string          `json:"transaction_id"`
	Fee           decimal.Decimal `json:"fee"` // in TRX
}

type BitcoinWithdrawResponse struct {
	TransactionID string          `json:"transaction_id"`
	Fee           decimal.Decimal `json:"fee"` // in BTC
}

type EthereumWalletBalanceRequest struct {
	Address     ethereum_address.EthereumAddress   `json:"address"`
	IncludeETH  bool                               `json:"includeETH"`
	ERC20Tokens []ethereum_address.EthereumAddress `json:"erc20Tokens"`
}

type EthereumWithdrawRequest struct {
	ContractAddress *ethereum_address.EthereumAddress `json:"contractAddress,omitempty"`
	PrivateKey      string                            `json:"privateKey"`
	ToAddress       ethereum_address.EthereumAddress  `json:"toAddress"`
	Amount          decimal.Decimal                   `json:"amount"`
}

type EthereumWithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
}

type SolanaWalletBalanceRequest struct {
	Address solana_address.SolanaAddress `json:"address"`
}

type SolanaWithdrawRequest struct {
	ExpandedPrivateKey string                       `json:"expanded_private_key"` // seed (32 bytes) + 32 bytes of public key
	ToAddress          solana_address.SolanaAddress `json:"to_address"`
	Amount             decimal.Decimal              `json:"amount"`
}

type XRPLWalletBalanceRequest struct {
	Address xrpl_address.Address `json:"address"`
}

type XRPLWithdrawRequest struct {
	PrivateKey string               `json:"private_key"`
	ToAddress  xrpl_address.Address `json:"to_address"`
	Amount     decimal.Decimal      `json:"amount"`
}

type XRPLWithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
}

type SolanaWithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
}

type SolanaGetWalletTransactionsRequest struct {
	WalletAddress       solana_address.SolanaAddress `json:"walletAddress"`
	LatestTransactionId *string                      `json:"latestTransactionId,omitempty"`
}

type SolanaGetWalletTransactionsResponse struct {
	Transactions []SolanaTransaction `json:"transactions"`
}

type SolanaTransaction struct {
	TxID      string          `json:"txid"`
	Address   string          `json:"address"`
	Fee       decimal.Decimal `json:"fee"`
	Amount    decimal.Decimal `json:"amount"`
	Timestamp int64           `json:"timestamp"`
}

type SolanaParseBlocksRequest struct {
	LatestBlockNumber *uint64                        `json:"latestBlockNumber,omitempty"`
	WalletAddresses   []solana_address.SolanaAddress `json:"walletAddresses"`
	TokenID           string                         `json:"tokenId"`
}

type BSCWalletBalanceRequest struct {
	Address     bsc_address.BSCAddress   `json:"address"`
	IncludeBNB  bool                     `json:"include_bnb"`
	BEP20Tokens []bsc_address.BSCAddress `json:"bep20_tokens"`
}

type BitcoinWalletBalanceRequest struct {
	Address bitcoin_address.BitcoinAddress `json:"address"`
}

type BitcoinWithdrawRequest struct {
	ToAddress bitcoin_address.BitcoinAddress `json:"to_address"`
	Amount    decimal.Decimal                `json:"amount"`
	UTXOs     []UTXO                         `json:"utxos"`
}

type UTXO struct {
	TransactionID     string          `json:"transaction_id"`
	Vout              int             `json:"vout"`
	PrivateKey        string          `json:"private_key"`
	Amount            decimal.Decimal `json:"amount"`
	ScriptPubKeyBytes []byte          `json:"script_pubkey_bytes"`
}

type ParseBlocksToken struct {
	ID        uuid.UUID        `json:"id"`
	TokenID   *string          `json:"tokenId,omitempty"`
	TokenType models.TokenType `json:"tokenType"`
}

type ParseBlocksRequest struct {
	LatestBlockNumber *uint64                    `json:"latestBlockNumber,omitempty"`
	WalletAddresses   []tron_address.TronAddress `json:"walletAddresses"`
	Tokens            []ParseBlocksToken         `json:"tokens"`
}

type BitcoinParseBlocksRequest struct {
	LatestBlockNumber *uint64                          `json:"latestBlockNumber,omitempty"`
	WalletAddresses   []bitcoin_address.BitcoinAddress `json:"walletAddresses"`
	UTXOs             map[string]struct{}              `json:"utxos"`
}

type ParseBlocksResponse struct {
	ParsedBlocksNumbers []uint64  `json:"parsedBlocksNumbers"`
	FailedBlocksNumbers []uint64  `json:"failedBlocksNumbers"`
	Deposits            []Deposit `json:"deposits"`
	LastBlockNumber     uint64    `json:"lastBlockNumber"`
}

type BitcoinParseBlocksResponse struct {
	ParsedBlocksNumbers []uint64             `json:"parsedBlocksNumbers"`
	FailedBlocksNumbers []uint64             `json:"failedBlocksNumbers"`
	Transactions        []BitcoinTransaction `json:"transactions"`
	LastBlockNumber     uint64               `json:"lastBlockNumber"`
}

type BitcoinTransaction struct {
	TxID string         `json:"txid"`
	Vin  []SpentUTXO    `json:"vin"`
	Vout []ReceivedUTXO `json:"vout"`
	Fee  float64        `json:"fee"`
	Time int64          `json:"time"`
}

type ReceivedUTXO struct {
	TxID            string `json:"txid"`
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	Vout            int    `json:"vout"`
	Scriptpubkeyhex string `json:"scriptpubkeyhex"`
}

type SpentUTXO struct {
	TxID string `json:"txid"`
	Vout int    `json:"vout"`
}

type Deposit struct {
	TxID      string `json:"txid"`
	TokenID   string `json:"tokenId"`
	ToAddress string `json:"toAddress"`
	Amount    string `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}
