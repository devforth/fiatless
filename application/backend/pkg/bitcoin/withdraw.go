package bitcoin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	bp_models "fiatless/internal/bp-models"
	"fiatless/internal/models"
	"fiatless/pkg/bitcoin/address"

	"bytes"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/mr-tron/base58"
	"github.com/shopspring/decimal"
)

// WithdrawProcessor builds and signs Bitcoin transactions using btcwallet's txauthor.
// It expects callers to provide UTXOs from the application's database.
type WithdrawProcessor struct {
	PrivateKey  *btcec.PrivateKey
	Network     models.BitcoinNetwork
	AddressType address.BitcoinAddressType
	// Change address to send any remainder to. For now, a placeholder can be used.
	ChangeAddress string
}

// NewWithdrawProcessor creates a new processor instance.
func NewWithdrawProcessor(priv *btcec.PrivateKey, network models.BitcoinNetwork, addrType address.BitcoinAddressType) *WithdrawProcessor {
	return &WithdrawProcessor{
		PrivateKey:    priv,
		Network:       network,
		AddressType:   addrType,
		ChangeAddress: "tb1pg5etccdtz0nqtrq6mgufrhfln2lx44w3llf3haytkvxte5qxw65slc3vfk",
	}
}

// CreateAndSign builds a transaction paying the specified amount to toAddr using provided UTXOs,
// adds change to the configured change address, signs inputs where supported, and returns the tx.
//
// Notes:
// - Fee rate is fixed to 1 sat/vB (i.e., 1000 sat/KB) for now.
// - Signing supports native SegWit P2WPKH and Taproot (P2TR) keypath inputs.
func (p *WithdrawProcessor) CreateAndSign(ctx context.Context, utxos []models.UTXO, toAddr address.BitcoinAddress, amount decimal.Decimal) (*wire.MsgTx, btcutil.Amount, error) {
	if len(utxos) == 0 {
		return nil, 0, errors.New("no UTXOs provided")
	}

	// Convert target amount to satoshis
	sats, err := decimalToSats(amount)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid amount: %w", err)
	}

	// Build primary output to recipient
	toScript, err := txscript.PayToAddrScript(toAddrRaw(toAddr, p.Network))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create recipient script: %w", err)
	}
	outputs := []*wire.TxOut{wire.NewTxOut(int64(sats), toScript)}

	// Prepare change address with placeholder
	changeAddr, err := decodeAddress(p.ChangeAddress, p.Network)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid change address: %w", err)
	}

	// Convert UTXOs to internal coins representation
	coins := convertToCoins(utxos)

	// Select coins and fee assuming 1 sat/vB
	selected, changeAmt, fee, selErr := p.selectCoinsAndFee(coins, outputs, changeAddr)
	if selErr != nil {
		return nil, 0, selErr
	}

	// Build transaction
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, c := range selected {
		tx.AddTxIn(wire.NewTxIn(&c.outPoint, nil, nil))
	}
	tx.AddTxOut(outputs[0])
	if changeAmt > 0 {
		changeScript, _ := txscript.PayToAddrScript(changeAddr)
		tx.AddTxOut(wire.NewTxOut(int64(changeAmt), changeScript))
	}

	// Sign inputs we can (P2WPKH and P2TR keypath)
	if err := p.signSupportedInputs(tx, selected); err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

// CreateAndSignWithPerInputKeys builds and signs a transaction using coin set that embeds
// per-input private keys (for multi-address input sets).
func (p *WithdrawProcessor) CreateAndSignWithPerInputKeys(ctx context.Context, coins []coin, toAddr address.BitcoinAddress, amount decimal.Decimal) (*wire.MsgTx, btcutil.Amount, error) {
	if len(coins) == 0 {
		return nil, 0, errors.New("no UTXOs provided")
	}

	sats, err := decimalToSats(amount)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid amount: %w", err)
	}

	toScript, err := txscript.PayToAddrScript(toAddrRaw(toAddr, p.Network))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create recipient script: %w", err)
	}
	outputs := []*wire.TxOut{wire.NewTxOut(int64(sats), toScript)}

	changeAddr, err := decodeAddress(p.ChangeAddress, p.Network)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid change address: %w", err)
	}

	selected, changeAmt, fee, selErr := p.selectCoinsAndFee(coins, outputs, changeAddr)
	if selErr != nil {
		return nil, 0, selErr
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	for _, c := range selected {
		tx.AddTxIn(wire.NewTxIn(&c.outPoint, nil, nil))
	}
	tx.AddTxOut(outputs[0])
	if changeAmt > 0 {
		changeScript, _ := txscript.PayToAddrScript(changeAddr)
		tx.AddTxOut(wire.NewTxOut(int64(changeAmt), changeScript))
	}

	if err := p.signSupportedInputs(tx, selected); err != nil {
		return nil, 0, err
	}

	return tx, fee, nil
}

// signSupportedInputs signs P2WPKH and P2TR keypath inputs using per-coin private key.
func (p *WithdrawProcessor) signSupportedInputs(tx *wire.MsgTx, coins []coin) error {
	for idx, txIn := range tx.TxIn {
		c, ok := findCoinByOutPoint(coins, txIn.PreviousOutPoint)
		if !ok {
			return fmt.Errorf("missing coin for input %d", idx)
		}

		if txscript.IsPayToWitnessPubKeyHash(c.pkScript) {
			// Verify priv matches pkScript's pubkeyhash
			pubBytes := c.privKey.PubKey().SerializeCompressed()
			computedHash := btcutil.Hash160(pubBytes)
			expectedHash := c.pkScript[2:22] // OP_0 OP_PUSHBYTES_20 [20-byte hash]
			if !bytes.Equal(computedHash, expectedHash) {
				return fmt.Errorf("private key does not match P2WPKH input %d scriptPubKey", idx)
			}

			// Proceed to sign
			prevFetcher := txscript.NewCannedPrevOutputFetcher(c.pkScript, int64(c.value))
			sigHashes := txscript.NewTxSigHashes(tx, prevFetcher)
			witness, err := txscript.WitnessSignature(tx, sigHashes, idx, int64(c.value), c.pkScript, txscript.SigHashAll, c.privKey, true)
			if err != nil {
				return fmt.Errorf("failed to create witness for input %d: %w", idx, err)
			}
			tx.TxIn[idx].Witness = witness
			continue
		}

		if txscript.IsPayToTaproot(c.pkScript) {
			expectedXOnly := c.pkScript[2:34] // OP_1 OP_PUSHBYTES_32 [32-byte x-only]
			// Cross-check using btcd's ComputeTaprootKeyNoScript (handles internal parity)
			btcdOutKey := txscript.ComputeTaprootKeyNoScript(c.privKey.PubKey())
			btcdXOnly := schnorr.SerializePubKey(btcdOutKey)
			if !bytes.Equal(btcdXOnly, expectedXOnly) {
				return fmt.Errorf("internal key does not derive this P2TR output (BIP86 mismatch)")
			}

			prevFetcher := txscript.NewCannedPrevOutputFetcher(c.pkScript, int64(c.value))
			sigHashes := txscript.NewTxSigHashes(tx, prevFetcher)
			// Manually compute taproot sighash and signature using tweaked private key
			sighash, err := txscript.CalcTaprootSignatureHash(sigHashes, txscript.SigHashDefault, tx, idx, prevFetcher)
			if err != nil {
				return fmt.Errorf("failed to calc taproot sighash for input %d: %w", idx, err)
			}
			// Use btcd's tweak (handles internal key parity per BIP341)
			tweakedPriv := txscript.TweakTaprootPrivKey(*c.privKey, []byte{})
			sig, err := schnorr.Sign(tweakedPriv, sighash)
			if err != nil {
				return fmt.Errorf("failed to sign taproot input %d: %w", idx, err)
			}
			tx.TxIn[idx].Witness = wire.TxWitness{sig.Serialize()}
			continue
		}

		return fmt.Errorf("unsupported script type for input %d", idx)
	}
	return nil
}

// Utilities

func decimalToSats(d decimal.Decimal) (btcutil.Amount, error) {
	satMultiplier := decimal.NewFromInt(100000000)
	v := d.Mul(satMultiplier).Round(0)
	if v.IsNegative() {
		return 0, errors.New("negative amount")
	}
	return btcutil.Amount(v.IntPart()), nil
}

func toParams(n models.BitcoinNetwork) *chaincfg.Params {
	switch n {
	case models.BitcoinMainnet:
		return &chaincfg.MainNetParams
	case models.BitcoinSignet:
		return &chaincfg.SigNetParams
	default:
		return &chaincfg.MainNetParams
	}
}

func decodeAddress(addr string, net models.BitcoinNetwork) (btcutil.Address, error) {
	return btcutil.DecodeAddress(addr, toParams(net))
}

func toAddrRaw(a address.BitcoinAddress, net models.BitcoinNetwork) btcutil.Address {
	// Re-decode to get btcutil.Address under the right net params
	// because address.BitcoinAddress holds btcutil.Address but network may differ
	aa, _ := btcutil.DecodeAddress(a.String(), toParams(net))
	return aa
}

// coin is an internal representation used by the InputSource
type coin struct {
	outPoint wire.OutPoint
	value    btcutil.Amount
	pkScript []byte
	privKey  *btcec.PrivateKey
}

func convertToCoins(utxos []models.UTXO) []coin {
	coins := make([]coin, 0, len(utxos))
	for _, u := range utxos {
		h, err := chainhash.NewHashFromStr(u.Transaction.TxID)
		if err != nil {
			continue
		}
		amt, _ := decimalToSats(u.Amount)
		coins = append(coins, coin{
			outPoint: wire.OutPoint{Hash: *h, Index: uint32(u.Vout)},
			value:    amt,
			pkScript: u.ScriptPubKeyBytes,
		})
	}
	return coins
}

func convertBpUtxosToCoins(utxos []bp_models.UTXO) ([]coin, error) {
	coins := make([]coin, 0, len(utxos))
	for _, u := range utxos {
		h, err := chainhash.NewHashFromStr(u.TransactionID)
		if err != nil {
			return nil, err
		}
		amt, err := decimalToSats(u.Amount)
		if err != nil {
			return nil, err
		}
		priv, err := parseBitcoinPrivKey(u.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		coins = append(coins, coin{
			outPoint: wire.OutPoint{Hash: *h, Index: uint32(u.Vout)},
			value:    amt,
			pkScript: u.ScriptPubKeyBytes,
			privKey:  priv,
		})
	}
	return coins, nil
}

// parseBitcoinPrivKey attempts multiple encodings: WIF, 64-hex, base58-raw32.
func parseBitcoinPrivKey(s string) (*btcec.PrivateKey, error) {
	// Try WIF
	if wif, err := btcutil.DecodeWIF(s); err == nil && wif != nil && wif.PrivKey != nil {
		return wif.PrivKey, nil
	}
	// Try hex
	if len(s) == 64 {
		if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
			if priv, _ := btcec.PrivKeyFromBytes(b); priv != nil {
				return priv, nil
			}
		}
	}
	// Try base58 raw 32-byte
	if b, err := base58.Decode(s); err == nil && len(b) == 32 {
		if priv, _ := btcec.PrivKeyFromBytes(b); priv != nil {
			return priv, nil
		}
	}
	return nil, errors.New("unsupported private key format (WIF, hex32, or base58-32 expected)")
}

// selectCoinsAndFee selects UTXOs and computes fee for 1 sat/vB. Returns selected coins,
// change amount (0 if none) and total fee.
func (p *WithdrawProcessor) selectCoinsAndFee(coins []coin, outputs []*wire.TxOut, changeAddr btcutil.Address) ([]coin, btcutil.Amount, btcutil.Amount, error) {
	var selected []coin
	var totalIn btcutil.Amount
	feeRate := btcutil.Amount(1) // 1 sat/vB

	// Precompute base lengths
	recipientLen := outputSerializedLen(outputs[0])
	changeScript, _ := txscript.PayToAddrScript(changeAddr)
	changeLen := 8 + varIntSerializeSize(uint64(len(changeScript))) + len(changeScript)

	recipientAmt := btcutil.Amount(outputs[0].Value)

	for _, c := range coins {
		selected = append(selected, c)
		totalIn += c.value

		vsizeTwoOut := estimateVSizeWithInputs(selected, recipientLen+changeLen)
		feeTwoOut := feeRate * btcutil.Amount(vsizeTwoOut)
		changeAmt := totalIn - recipientAmt - feeTwoOut
		if changeAmt >= 0 {
			if changeAmt < btcutil.Amount(330) { // dust threshold approx
				vsizeOneOut := estimateVSizeWithInputs(selected, recipientLen)
				feeOneOut := feeRate * btcutil.Amount(vsizeOneOut)
				if totalIn >= recipientAmt+feeOneOut {
					return selected, 0, feeOneOut, nil
				}
			} else {
				return selected, changeAmt, feeTwoOut, nil
			}
		}
	}
	return nil, 0, 0, errors.New("insufficient funds")
}

func estimateVSizeWithInputs(inputs []coin, outputsLen int) int {
	// version(4) + marker(1) + flag(1) + locktime(4) + varint inputs(1) + varint outputs(1)
	base := 4 + 1 + 1 + 4 + 1 + 1
	// Sum per-input vbytes based on script type
	inVBytes := 0
	for _, c := range inputs {
		switch {
		case txscript.IsPayToTaproot(c.pkScript):
			// Approx. P2TR keypath input virtual size
			inVBytes += 58
		case txscript.IsPayToWitnessPubKeyHash(c.pkScript):
			// Approx. P2WPKH input virtual size
			inVBytes += 68
		default:
			// Unsupported input types for our signer path; be conservative
			inVBytes += 148
		}
	}
	// Outputs length already excludes witness
	outBytes := outputsLen
	return base + inVBytes + outBytes
}

func varIntSerializeSize(val uint64) int {
	switch {
	case val < 0xfd:
		return 1
	case val <= 0xffff:
		return 3
	case val <= 0xffffffff:
		return 5
	default:
		return 9
	}
}

func outputSerializedLen(out *wire.TxOut) int {
	pkLen := len(out.PkScript)
	return 8 + varIntSerializeSize(uint64(pkLen)) + pkLen
}

func findCoinByOutPoint(coins []coin, op wire.OutPoint) (coin, bool) {
	for _, c := range coins {
		if c.outPoint == op {
			return c, true
		}
	}
	return coin{}, false
}
