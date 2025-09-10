package replay

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	log "github.com/xlab/suplog"
)

// GasFuzzingConfig defines configuration options for gas limit fuzzing
type GasFuzzingConfig struct {
	// Enabled controls whether gas limit fuzzing is active
	Enabled bool

	// Strategy defines the fuzzing approach: "random", "boundary"
	Strategy string

	// GasLimitMultiplierMin minimum multiplier for gas limit (default: 0.01)
	GasLimitMultiplierMin float64

	// GasLimitMultiplierMax maximum multiplier for gas limit (default: 1.0)
	GasLimitMultiplierMax float64

	// Seed for deterministic fuzzing (0 for random)
	Seed int64

	// FuzzPercentage percentage of transactions to fuzz (0-100)
	FuzzPercentage int

	// VerboseLogging enables detailed before/after transaction logging
	VerboseLogging bool
}

// GasFuzzer handles gas limit fuzzing for transactions during replay
type GasFuzzer struct {
	config    GasFuzzingConfig
	rng       *rand.Rand
	txCounter uint64
	logger    log.Logger
}

// NewGasFuzzer creates a new gas fuzzer instance
func NewGasFuzzer(config GasFuzzingConfig) *GasFuzzer {
	if !config.Enabled {
		return nil
	}

	// Set default values if not specified
	if config.GasLimitMultiplierMin == 0 {
		config.GasLimitMultiplierMin = 0.5
	}
	if config.GasLimitMultiplierMax == 0 {
		config.GasLimitMultiplierMax = 5.0
	}
	if config.FuzzPercentage == 0 {
		config.FuzzPercentage = 100 // Default to fuzzing all transactions
	}
	if config.Strategy == "" {
		config.Strategy = "random"
	}

	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &GasFuzzer{
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
		logger: log.WithField("component", "gas-fuzzer"),
	}
}

// FuzzTransaction applies gas fuzzing to the given transaction bytes
func (f *GasFuzzer) FuzzTransaction(txBytes []byte, client chain.Client) ([]byte, error) {
	if f == nil || !f.config.Enabled {
		return txBytes, nil
	}

	// Decide whether to fuzz this transaction based on FuzzPercentage
	if f.rng.Intn(100) >= f.config.FuzzPercentage {
		return txBytes, nil
	}

	f.txCounter++

	// Decode transaction
	txDecoder := client.TxConfig().TxDecoder()
	tx, err := txDecoder(txBytes)
	if err != nil {
		f.logger.WithError(err).Warning("Failed to decode transaction for fuzzing, skipping")
		return txBytes, nil
	}

	// Extract gas information
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		f.logger.Warning("Transaction does not implement FeeTx interface, skipping fuzzing")
		return txBytes, nil
	}

	originalGasLimit := feeTx.GetGas()
	originalFee := feeTx.GetFee()

	// Apply fuzzing strategy
	newGasLimit, err := f.applyFuzzingStrategy(originalGasLimit)
	if err != nil {
		f.logger.WithError(err).Warning("Failed to apply fuzzing strategy, using original values")
		return txBytes, nil
	}

	// Create new transaction with fuzzed gas values
	fuzzedTx, err := f.createFuzzedTransaction(tx, newGasLimit, originalFee, client)
	if err != nil {
		f.logger.WithError(err).Warning("Failed to create fuzzed transaction, using original")
		return txBytes, nil
	}

	return client.Encode(fuzzedTx), nil
}

// applyFuzzingStrategy applies the configured fuzzing strategy to gas limit
func (f *GasFuzzer) applyFuzzingStrategy(gasLimit uint64) (uint64, error) {
	switch f.config.Strategy {
	case "random":
		return f.randomFuzzing(gasLimit)
	case "boundary":
		return f.boundaryFuzzing(gasLimit)
	default:
		return f.randomFuzzing(gasLimit)
	}
}

// randomFuzzing applies random gas limit multiplier within configured range
func (f *GasFuzzer) randomFuzzing(gasLimit uint64) (uint64, error) {
	// Random gas limit multiplier
	gasMultiplier := f.config.GasLimitMultiplierMin +
		f.rng.Float64()*(f.config.GasLimitMultiplierMax-f.config.GasLimitMultiplierMin)
	newGasLimit := uint64(float64(gasLimit) * gasMultiplier)

	return newGasLimit, nil
}

// boundaryFuzzing tests extreme gas limit values (min/max multipliers)
func (f *GasFuzzer) boundaryFuzzing(gasLimit uint64) (uint64, error) {
	// Alternate between min and max values based on transaction counter
	useMin := f.txCounter%2 == 0

	var gasMultiplier float64
	if useMin {
		gasMultiplier = f.config.GasLimitMultiplierMin
	} else {
		gasMultiplier = f.config.GasLimitMultiplierMax
	}

	newGasLimit := uint64(float64(gasLimit) * gasMultiplier)

	return newGasLimit, nil
}

// createFuzzedTransaction creates a new transaction with modified gas limit
func (f *GasFuzzer) createFuzzedTransaction(
	originalTx sdk.Tx,
	gasLimit uint64,
	fee sdk.Coins,
	client chain.Client,
) (authsigning.Tx, error) {
	// Create a new transaction builder
	txBuilder := client.TxConfig().NewTxBuilder()

	// Copy messages from original transaction
	msgs := originalTx.GetMsgs()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, errors.Wrap(err, "failed to set messages in fuzzed transaction")
	}

	// Set fuzzed gas values
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fee)

	// Copy fee granter and fee payer information (critical for authz transactions)
	if feeTx, ok := originalTx.(sdk.FeeTx); ok {
		if feeGranter := feeTx.FeeGranter(); len(feeGranter) > 0 {
			txBuilder.SetFeeGranter(sdk.AccAddress(feeGranter))
		}

		if feePayer := feeTx.FeePayer(); len(feePayer) > 0 {
			txBuilder.SetFeePayer(sdk.AccAddress(feePayer))
		}
	}

	// Copy other transaction attributes
	if memTx, ok := originalTx.(sdk.TxWithMemo); ok {
		txBuilder.SetMemo(memTx.GetMemo())
	}

	if timeoutTx, ok := originalTx.(sdk.TxWithTimeoutHeight); ok {
		txBuilder.SetTimeoutHeight(timeoutTx.GetTimeoutHeight())
	}

	// Copy signatures as-is from the original transaction
	if sigTx, ok := originalTx.(authsigning.SigVerifiableTx); ok {
		signatures, err := sigTx.GetSignaturesV2()
		if err != nil {
			f.logger.WithError(err).Warning("Failed to get signatures from original transaction, proceeding with unsigned transaction")
		} else if len(signatures) > 0 {
			if err := txBuilder.SetSignatures(signatures...); err != nil {
				f.logger.WithError(err).Warning("Failed to set signatures on fuzzed transaction, proceeding with unsigned transaction")
			}
		} else {
			f.logger.Debug("Original transaction has no signatures to copy")
		}
	}

	return txBuilder.GetTx(), nil
}

// ExtractTransactionInfo decodes transaction bytes and extracts gas information for logging
func ExtractTransactionInfo(txBytes []byte, client chain.Client, txType string) log.Fields {
	if len(txBytes) == 0 {
		return nil
	}

	// Decode transaction
	txDecoder := client.TxConfig().TxDecoder()
	tx, err := txDecoder(txBytes)
	if err != nil {
		return log.Fields{
			"tx_type":      txType,
			"decode_error": err.Error(),
			"tx_size":      len(txBytes),
		}
	}

	fields := log.Fields{
		"tx_type": txType,
		"tx_size": len(txBytes),
	}

	// Extract gas information
	if feeTx, ok := tx.(sdk.FeeTx); ok {
		fields["gas_limit"] = feeTx.GetGas()

		fee := feeTx.GetFee()
		fields["fee_coins"] = fee.String()

		// Calculate total fee amount
		totalFee := sdk.NewCoins()
		for _, coin := range fee {
			totalFee = totalFee.Add(coin)
		}
		fields["total_fee"] = totalFee.String()

		// Extract fee granter and fee payer (important for authz transactions)
		if feeGranter := feeTx.FeeGranter(); len(feeGranter) > 0 {
			fields["fee_granter"] = sdk.AccAddress(feeGranter).String()
		}

		if feePayer := feeTx.FeePayer(); len(feePayer) > 0 {
			fields["fee_payer"] = sdk.AccAddress(feePayer).String()
		}
	} else {
		fields["gas_limit"] = "unknown (not FeeTx)"
		fields["total_fee"] = "unknown (not FeeTx)"
	}

	// Extract messages information
	msgs := tx.GetMsgs()
	fields["msg_count"] = len(msgs)
	if len(msgs) > 0 {
		msgTypes := make([]string, len(msgs))
		for i, msg := range msgs {
			msgTypes[i] = sdk.MsgTypeURL(msg)
		}
		fields["msg_types"] = strings.Join(msgTypes, ",")
	}

	// Extract memo if available
	if memoTx, ok := tx.(sdk.TxWithMemo); ok {
		memo := memoTx.GetMemo()
		if memo != "" {
			fields["memo"] = memo
		}
	}

	// Extract signature information and signer address
	if sigTx, ok := tx.(authsigning.SigVerifiableTx); ok {
		signatures, err := sigTx.GetSignaturesV2()
		if err != nil {
			fields["signature_error"] = err.Error()
		} else {
			fields["signature_count"] = len(signatures)
			if len(signatures) > 0 {
				// Get signer address from first signature
				if pubKey := signatures[0].PubKey; pubKey != nil {
					signerAddr := sdk.AccAddress(pubKey.Address())
					fields["sender"] = signerAddr.String()
					fields["first_sig_pubkey"] = fmt.Sprintf("%x", pubKey.Bytes()[:8])
				}
				fields["first_sig_sequence"] = signatures[0].Sequence
			}
		}
	} else {
		fields["signature_count"] = "unknown (not SigVerifiableTx)"
	}

	// Note: Extension options extraction is not supported in this SDK version

	// Calculate transaction hash for identification
	hash := sha256.Sum256(txBytes)
	fields["tx_hash"] = fmt.Sprintf("%x", hash[:8]) // First 8 bytes for readability

	return fields
}
