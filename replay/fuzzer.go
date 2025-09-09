package replay

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/pkg/errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	log "github.com/xlab/suplog"
)

// GasFuzzingConfig defines configuration options for gas fuzzing
type GasFuzzingConfig struct {
	// Enabled controls whether gas fuzzing is active
	Enabled bool

	// Strategy defines the fuzzing approach: "random", "boundary"
	Strategy string

	// GasLimitMultiplierMin minimum multiplier for gas limit (default: 0.5)
	GasLimitMultiplierMin float64

	// GasLimitMultiplierMax maximum multiplier for gas limit (default: 5.0)
	GasLimitMultiplierMax float64

	// GasPriceMultiplierMin minimum multiplier for gas price (default: 0.1)
	GasPriceMultiplierMin float64

	// GasPriceMultiplierMax maximum multiplier for gas price (default: 10.0)
	GasPriceMultiplierMax float64

	// Seed for deterministic fuzzing (0 for random)
	Seed int64

	// FuzzPercentage percentage of transactions to fuzz (0-100)
	FuzzPercentage int

	// VerboseLogging enables detailed before/after transaction logging
	VerboseLogging bool
}

// GasFuzzer handles gas value fuzzing for transactions during replay
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
	if config.GasPriceMultiplierMin == 0 {
		config.GasPriceMultiplierMin = 0.1
	}
	if config.GasPriceMultiplierMax == 0 {
		config.GasPriceMultiplierMax = 10.0
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
	newGasLimit, newFee, err := f.applyFuzzingStrategy(originalGasLimit, originalFee)
	if err != nil {
		f.logger.WithError(err).Warning("Failed to apply fuzzing strategy, using original values")
		return txBytes, nil
	}

	// Create new transaction with fuzzed gas values
	fuzzedTx, err := f.createFuzzedTransaction(tx, newGasLimit, newFee, client)
	if err != nil {
		f.logger.WithError(err).Warning("Failed to create fuzzed transaction, using original")
		return txBytes, nil
	}

	// Calculate multipliers for logging
	gasLimitMultiplier := float64(newGasLimit) / float64(originalGasLimit)

	f.logger.WithFields(log.Fields{
		"strategy":             f.config.Strategy,
		"tx_counter":           f.txCounter,
		"original_gas_limit":   originalGasLimit,
		"fuzzed_gas_limit":     newGasLimit,
		"gas_limit_multiplier": fmt.Sprintf("%.3f", gasLimitMultiplier),
		"original_fee":         originalFee.String(),
		"fuzzed_fee":           newFee.String(),
	}).Debug("💫 Gas fuzzer: Transaction modified")

	return client.Encode(fuzzedTx), nil
}

// applyFuzzingStrategy applies the configured fuzzing strategy to gas values
func (f *GasFuzzer) applyFuzzingStrategy(gasLimit uint64, fee sdk.Coins) (uint64, sdk.Coins, error) {
	switch f.config.Strategy {
	case "random":
		return f.randomFuzzing(gasLimit, fee)
	case "boundary":
		return f.boundaryFuzzing(gasLimit, fee)
	default:
		return f.randomFuzzing(gasLimit, fee)
	}
}

// randomFuzzing applies random multipliers within configured ranges
func (f *GasFuzzer) randomFuzzing(gasLimit uint64, fee sdk.Coins) (uint64, sdk.Coins, error) {
	// Random gas limit multiplier
	gasMultiplier := f.config.GasLimitMultiplierMin +
		f.rng.Float64()*(f.config.GasLimitMultiplierMax-f.config.GasLimitMultiplierMin)
	newGasLimit := uint64(float64(gasLimit) * gasMultiplier)

	// Random gas price multiplier
	priceMultiplier := f.config.GasPriceMultiplierMin +
		f.rng.Float64()*(f.config.GasPriceMultiplierMax-f.config.GasPriceMultiplierMin)

	newFee := f.multiplyCoins(fee, priceMultiplier)

	return newGasLimit, newFee, nil
}

// boundaryFuzzing tests extreme values (min/max multipliers)
func (f *GasFuzzer) boundaryFuzzing(gasLimit uint64, fee sdk.Coins) (uint64, sdk.Coins, error) {
	// Alternate between min and max values based on transaction counter
	useMin := f.txCounter%2 == 0

	var gasMultiplier, priceMultiplier float64
	if useMin {
		gasMultiplier = f.config.GasLimitMultiplierMin
		priceMultiplier = f.config.GasPriceMultiplierMin
	} else {
		gasMultiplier = f.config.GasLimitMultiplierMax
		priceMultiplier = f.config.GasPriceMultiplierMax
	}

	newGasLimit := uint64(float64(gasLimit) * gasMultiplier)
	newFee := f.multiplyCoins(fee, priceMultiplier)

	return newGasLimit, newFee, nil
}

// multiplyCoins multiplies all coin amounts by the given multiplier
func (f *GasFuzzer) multiplyCoins(coins sdk.Coins, multiplier float64) sdk.Coins {
	result := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		// Simple multiplication with precision handling
		newAmount := coin.Amount.MulRaw(int64(multiplier * 1000000)).QuoRaw(1000000)
		if newAmount.IsZero() {
			newAmount = sdkmath.NewInt(1)
		}
		result[i] = sdk.NewCoin(coin.Denom, newAmount)
	}
	return result
}

// createFuzzedTransaction creates a new transaction with modified gas values
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

	// Copy other transaction attributes
	if memTx, ok := originalTx.(sdk.TxWithMemo); ok {
		txBuilder.SetMemo(memTx.GetMemo())
	}

	if timeoutTx, ok := originalTx.(sdk.TxWithTimeoutHeight); ok {
		txBuilder.SetTimeoutHeight(timeoutTx.GetTimeoutHeight())
	}

	// Copy signatures (this is tricky as signatures may become invalid)
	// For replay scenarios, we'll create an unsigned transaction
	// The signature validation will be handled by the receiving chain

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

	// Calculate transaction hash for identification
	hash := sha256.Sum256(txBytes)
	fields["tx_hash"] = fmt.Sprintf("%x", hash[:8]) // First 8 bytes for readability

	return fields
}
