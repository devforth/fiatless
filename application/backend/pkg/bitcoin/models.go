package bitcoin

type BitcoinBlock struct {
	Hash         string               `json:"hash"`
	Height       int64                `json:"height"`
	Time         int64                `json:"time"`
	Transactions []BitcoinTransaction `json:"tx"`
}

type BitcoinTransaction struct {
	TxID string                      `json:"txid"`
	Vin  *[]BitcoinTransactionInput  `json:"vin,omitempty"`
	Vout *[]BitcoinTransactionOutput `json:"vout,omitempty"`
	Fee  *float64                    `json:"fee,omitempty"`
}

type BitcoinTransactionInput struct {
	TxID *string `json:"txid,omitempty"`
	Vout *int    `json:"vout,omitempty"`
}

type BitcoinTransactionOutput struct {
	Value        *float64                        `json:"value,omitempty"`
	N            *int                            `json:"n,omitempty"`
	ScriptPubKey *BitcoinTransactionScriptPubKey `json:"scriptPubKey,omitempty"`
}

type BitcoinTransactionScriptPubKey struct {
	Address *string `json:"address,omitempty"`
	Hex     *string `json:"hex,omitempty"`
}
