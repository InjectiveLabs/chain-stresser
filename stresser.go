package stresser

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/dottedmag/parallel"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"github.com/xlab/catcher"
	"github.com/xlab/pace"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/InjectiveLabs/chain-stresser/v2/payload"
	"github.com/InjectiveLabs/chain-stresser/v2/pkg/ratelimit"
	"github.com/InjectiveLabs/chain-stresser/v2/replay"
)

// StressConfig is the config for stress runner
type StressConfig struct {
	// ChainID (Cosmos) of the chain to connect to
	ChainID string

	//  EthChainID (EIP-155) of the EVM chain to connect to
	EthChainID int64

	// MinGasPrice to use for sending transactions
	MinGasPrice string

	// RPC address of the node to connect to
	NodeAddress string

	// GRPC address of the node to connect to
	GRPCAddress string

	// Account privkeys to use for sending transactions
	Accounts []chain.Secp256k1PrivateKey

	// NumOfTransactions to send per account
	NumOfTransactions int

	// AwaitTxConfirmation to wait for transaction to be included in a block
	AwaitTxConfirmation bool

	// GasFuzzing configuration for gas value fuzzing during replay
	GasFuzzing replay.GasFuzzingConfig

	// RateLimit configuration for controlling transaction throughput
	RateLimit ratelimit.Config
}

// maxParallelInitialTxsBroadcasts is the maximum number of initial txs to broadcast in parallel.
// this is internal to the stresser and not configurable for now.
const maxParallelInitialTxsBroadcasts = 8

// handleFallbackBroadcast attempts to broadcast the original transaction when fuzzed transaction fails
func handleFallbackBroadcast(
	ctx context.Context,
	originalTxBytes []byte,
	broadcastFunc ratelimit.BroadcastFunc,
	fuzzedError string,
	logger log.Logger,
) (txHash string, err error, wasSuccessful bool) {
	logger.WithFields(log.Fields{
		"fuzzed_error": fuzzedError,
	}).Debug("🔄 Fuzzed transaction failed with out of gas error, retrying with original")

	fallbackTxHash, fallbackErr := broadcastFunc(ctx, originalTxBytes)
	if fallbackErr != nil {
		logger.WithFields(log.Fields{
			"fuzzed_error":   fuzzedError,
			"original_error": fallbackErr.Error(),
		}).Debug("🚫 Both fuzzed and original transactions failed")
		return "", fallbackErr, false
	}

	if fallbackTxHash == "" {
		return "", fmt.Errorf("fallback transaction returned empty hash"), false
	}

	logger.WithFields(log.Fields{
		"original_tx_hash": fallbackTxHash,
	}).Debug("✅ Original transaction succeeded after fuzzed failure")
	return fallbackTxHash, nil, true
}

// categorizeError categorizes broadcast errors into different types
func categorizeError(errMsg string, timeoutErrors, sequenceErrors, outOfGasErrors, otherErrors *int, logger log.Logger) {
	if strings.Contains(errMsg, "tx timeout height") {
		*timeoutErrors++
	} else if strings.Contains(errMsg, "account sequence mismatch") {
		*sequenceErrors++
	} else if strings.Contains(errMsg, "out of gas") {
		*outOfGasErrors++
	} else {
		*otherErrors++
	}
}

// fuzzTransactions applies gas fuzzing to a transaction if enabled
func fuzzTransactions(
	txBytes []byte,
	gasFuzzer *replay.GasFuzzer,
	accountClient chain.Client,
	config StressConfig,
	logger log.Logger,
) ([]byte, bool) {
	if gasFuzzer == nil {
		return txBytes, false
	}

	var originalTxInfo, fuzzedTxInfo log.Fields

	// Log original transaction details if verbose logging is enabled
	if config.GasFuzzing.VerboseLogging {
		originalTxInfo = replay.ExtractTransactionInfo(txBytes, accountClient, "ORIGINAL")
		if originalTxInfo != nil {
			logger.WithFields(originalTxInfo).Debug("📋 Transaction before fuzzing")
		}
	}

	originalSize := len(txBytes)
	fuzzedTxBytes, fuzzErr := gasFuzzer.FuzzTransaction(txBytes, accountClient)
	if fuzzErr != nil {
		logger.WithError(fuzzErr).Warning("Gas fuzzing failed, using original transaction")
		return txBytes, false
	}

	// Check if transaction was actually modified
	if len(fuzzedTxBytes) == originalSize && string(fuzzedTxBytes) == string(txBytes) {
		return txBytes, false
	}

	// Log fuzzed transaction details if verbose logging is enabled
	if config.GasFuzzing.VerboseLogging {
		fuzzedTxInfo = replay.ExtractTransactionInfo(fuzzedTxBytes, accountClient, "FUZZED")
		if fuzzedTxInfo != nil {
			logger.WithFields(fuzzedTxInfo).Debug("🎯 Transaction after fuzzing")
		}

		// Log comparison
		if originalTxInfo != nil && fuzzedTxInfo != nil {
			logger.WithFields(log.Fields{
				"gas_limit_change": fmt.Sprintf("%v → %v", originalTxInfo["gas_limit"], fuzzedTxInfo["gas_limit"]),
				"tx_hash_change":   fmt.Sprintf("%v → %v", originalTxInfo["tx_hash"], fuzzedTxInfo["tx_hash"]),
				"strategy":         config.GasFuzzing.Strategy,
			}).Debug("🔥 Gas limit fuzzing applied successfully")
		}
	}

	return fuzzedTxBytes, true
}

func Stress(
	ctx context.Context,
	config StressConfig,
	txProvider payload.TxProvider,
) error {
	logger := log.WithField("bench", txProvider.Name())

	if config.EthChainID == 0 {
		logger.WithFields(log.Fields{
			"chainID":    config.ChainID,
			"ethChainID": config.EthChainID,
		}).Fatal("❌ EthChainID is required")
	}

	if config.ChainID == "" {
		logger.WithFields(log.Fields{
			"chainID":    config.ChainID,
			"ethChainID": config.EthChainID,
		}).Fatal("❌ ChainID is required")
	}

	client := chain.NewClient(config.ChainID, config.NodeAddress, config.GRPCAddress)
	stopMempoolReporter := startMempoolReporter(ctx, client, logger)
	defer stopMempoolReporter()

	startTs := time.Now()
	signedTxPace := pace.New("signed tx", 10*time.Second, NewPaceReporter(logger))
	getAccountNumberSequencePace := pace.New("sequence fetched", 10*time.Second, NewPaceReporter(logger))
	broadcastTxPace := pace.New("sent tx", 10*time.Second, NewPaceReporter(logger))

	// Create pace meters for each provider name
	providerPaceMeters := make(map[string]pace.Pace)
	providerPaceMetersMux := new(sync.Mutex)
	getOrCreateProviderPace := func(providerName string) pace.Pace {
		providerPaceMetersMux.Lock()
		defer providerPaceMetersMux.Unlock()

		if paceMeter, exists := providerPaceMeters[providerName]; exists {
			return paceMeter
		}

		// Create a new pace meter for this provider
		paceMeter := pace.New(fmt.Sprintf("signed tx [%s]", providerName), 10*time.Second, NewPaceReporter(logger))
		providerPaceMeters[providerName] = paceMeter
		return paceMeter
	}

	numOfAccounts := len(config.Accounts)
	logger.WithFields(log.Fields{
		"num": numOfAccounts * config.NumOfTransactions,
	}).Info("Preparing signed transactions. Please wait ⏳")

	var signedTxs [][][]byte
	var initialAccountSequences []uint64
	txQueue := make(chan payload.Tx, 1000000)
	txSignedQueue := make(chan payload.Tx, 1000000)

	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {

		for n := 0; n < runtime.NumCPU(); n++ {
			spawn(fmt.Sprintf("signer-%d", n), parallel.Continue, func(ctx context.Context) error {
				defer catcher.Catch(
					catcher.RecvLog(true),
					catcher.RecvDie(1, true),
				)

				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case tx, ok := <-txQueue:
						if !ok {
							return nil
						}

						signedTx, err := txProvider.BuildAndSignTx(
							client,
							tx,
						)
						orPanic(err)

						select {
						case <-ctx.Done():
							return ctx.Err()

						case txSignedQueue <- signedTx:
							providerName := tx.Provider().Name()
							providerPace := getOrCreateProviderPace(providerName)
							providerPace.Step(1)
						}
					}
				}
			})
		}

		spawn("generate", parallel.Continue, func(ctx context.Context) error {
			defer func() {
				getAccountNumberSequencePace.Pause()
			}()

			defer catcher.Catch(
				catcher.RecvLog(true),
				catcher.RecvDie(1, true),
			)

			if len(config.Accounts) == 0 {
				return errors.New("empty accounts list")
			} else {
				// this ensures that the state required for benchmark is correctly initialized
				// for EVM transactions this usually deploys a smart contract. We can do it for each account
				// if some account state needs to be initialized as well.
				orPanic(createAndBroadcastInitialTxs(
					ctx,
					logger,
					signedTxPace,
					getAccountNumberSequencePace,
					broadcastTxPace,
					client,
					txProvider,
					config.Accounts,
				))
			}

			initialAccountSequencesMux := new(sync.Mutex)
			initialAccountSequences = make([]uint64, numOfAccounts)
			pool := workerpool.New(runtime.NumCPU())

			for fromIdx := 0; fromIdx < numOfAccounts; fromIdx++ {
				fromPrivateKey := config.Accounts[fromIdx]
				accAddress := fromPrivateKey.AccAddress()

				pool.Submit(func() {
					defer catcher.Catch(
						catcher.RecvLog(true),
						catcher.RecvDie(1, true),
					)

					accNum, accSeq, err := getAccountNumberSequence(ctx, client, accAddress)
					if err != nil {
						err = errors.Wrap(err, "❌ Fetching account number and sequence failed")
						logger.WithFields(log.Fields{
							"accIdx":  fromIdx,
							"address": accAddress,
						}).WithError(err).Fatalln("❌ Fetching account number and sequence failed")

						return
					}

					initialAccountSequencesMux.Lock()
					initialAccountSequences[fromIdx] = accSeq
					initialAccountSequencesMux.Unlock()
					getAccountNumberSequencePace.Step(1)

					txRequest := payload.TxRequest{
						Keys: config.Accounts,

						From: chain.Account{
							Name:     fmt.Sprintf("sender-%d", fromIdx),
							Key:      fromPrivateKey,
							Number:   accNum,
							Sequence: accSeq,
						},

						FromIdx: fromIdx,
					}

					for txIdx := 0; txIdx < config.NumOfTransactions; txIdx++ {
						txRequest.TxIdx = txIdx

						tx, err := txProvider.GenerateTx(txRequest)
						orPanic(err)

						select {
						case <-ctx.Done():
							logger.WithFields(log.Fields{
								"fromIdx": fromIdx,
								"txIdx":   txIdx,
							}).Fatalln("❌ Context ended prematurely")

							return
						case txQueue <- tx:
						}

						txRequest.From.Sequence++
					}
				})
			}

			pool.StopWait()
			return nil
		})

		spawn("collect", parallel.Exit, func(ctx context.Context) error {
			defer func() {
				signedTxPace.Pause()
				// Pause all provider-specific pace meters
				providerPaceMetersMux.Lock()
				for _, paceMeter := range providerPaceMeters {
					paceMeter.Pause()
				}
				providerPaceMetersMux.Unlock()
			}()

			defer catcher.Catch(
				catcher.RecvLog(true),
				catcher.RecvDie(1, true),
			)

			signedTxs = make([][][]byte, numOfAccounts)
			for i := 0; i < numOfAccounts; i++ {
				signedTxs[i] = make([][]byte, config.NumOfTransactions)
			}

			for i := 0; i < numOfAccounts; i++ {
				for j := 0; j < config.NumOfTransactions; j++ {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case txSigned := <-txSignedQueue:
						signedTxs[txSigned.FromIdx()][txSigned.TxIdx()] = txSigned.Bytes()
						signedTxPace.Step(1)
					}
				}
			}

			return nil
		})

		return nil
	})
	if err != nil {
		return err
	}

	logger.WithFields(log.Fields{
		"elapsed": time.Since(startTs),
	}).Infof("Transactions prepared 🙌")

	startTs = time.Now()

	logger.Info("Broadcasting transactions 🚀")

	// Create broadcast function with optional rate limiting
	broadcastFunc, err := buildBroadcastClient(config)
	if err != nil {
		return errors.Wrap(err, "failed to build broadcast client")
	} else if config.RateLimit.IsEnabled() {
		logger.WithFields(log.Fields{
			"tps_limit":        config.RateLimit.TxPerSecond,
			"bytes_per_second": config.RateLimit.BytesPerSecond,
			"gas_per_second":   config.RateLimit.GasPerSecond,
			"burst_size":       config.RateLimit.Burst.Size,
		}).Info("Rate limiter enabled 🐢")
	}

	if err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("accounts", parallel.Exit, func(ctx context.Context) error {
			return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
				for accountIdx, accountTxs := range signedTxs {
					accountTxs := accountTxs
					accountIdx := accountIdx

					initialSequence := initialAccountSequences[accountIdx]

					spawn(fmt.Sprintf("account-%d", accountIdx), parallel.Continue, func(ctx context.Context) error {
						defer catcher.Catch(
							catcher.RecvLog(true),
							catcher.RecvDie(1, true),
						)

						for txIndex := 0; txIndex < config.NumOfTransactions; {
							tx := accountTxs[txIndex]

							txHash, err := broadcastFunc(ctx, tx)
							if err != nil {
								if expectedAccSeq, ok := chain.IsSequenceError(err); ok {
									logger.WithError(err).WithFields(log.Fields{
										"accIndex":           accountIdx,
										"txIndex":            txIndex,
										"initialAccSequence": initialSequence,
										"expectedSequence":   expectedAccSeq,
										"newSequence":        int(expectedAccSeq - initialSequence),
									}).Debug("⚠️ Tx broadcasting failed, trying suggested sequence")

									txIndex = int(expectedAccSeq - initialSequence)
									continue
								}

								err = errors.Wrap(err, "⚠️ Tx broadcasting error")
								return err
							}

							broadcastTxPace.Step(1)
							logger.WithFields(log.Fields{
								"txHash": txHash,
							}).Debug("✅ Tx broadcasted")

							txIndex++
						}

						return nil
					})
				}

				return nil
			})
		})

		return nil
	}); err != nil {
		return err
	}

	broadcastTxPace.Pause()
	logger.WithFields(log.Fields{
		"broadcastDuration": time.Since(startTs),
	}).Info("Benchmark done 🎉")

	return nil
}

// StressReplay is used to replay raw transactions from a remote chain
func StressReplay(
	ctx context.Context,
	config StressConfig,
	txProvider payload.TxProvider,
) error {
	logger := log.WithField("bench", txProvider.Name())

	// Create broadcast function with optional rate limiting
	broadcastFunc, err := buildBroadcastClient(config)
	if err != nil {
		return fmt.Errorf("failed to create broadcast function: %w", err)
	}

	if config.RateLimit.IsEnabled() {
		logger.WithFields(log.Fields{
			"tps_limit":        config.RateLimit.TxPerSecond,
			"bytes_per_second": config.RateLimit.BytesPerSecond,
			"gas_per_second":   config.RateLimit.GasPerSecond,
			"burst_size":       config.RateLimit.Burst.Size,
		}).Info("Rate limiter enabled for replay 🐢")

	}

	// Create client for gas fuzzing (separate from broadcast function)
	accountClient := chain.NewClient(config.ChainID, config.NodeAddress, config.GRPCAddress)
	stopMempoolReporter := startMempoolReporter(ctx, accountClient, logger)
	defer stopMempoolReporter()
	broadcastTxPace := pace.New("sent tx", 10*time.Second, NewPaceReporter(logger))

	// Initialize gas fuzzer if enabled
	gasFuzzer := replay.NewGasFuzzer(config.GasFuzzing)
	if gasFuzzer != nil {
		logger.WithFields(log.Fields{
			"fuzzing_enabled":          true,
			"fuzzing_strategy":         config.GasFuzzing.Strategy,
			"fuzz_percentage":          config.GasFuzzing.FuzzPercentage,
			"gas_limit_multiplier_min": config.GasFuzzing.GasLimitMultiplierMin,
			"gas_limit_multiplier_max": config.GasFuzzing.GasLimitMultiplierMax,
		}).Info("Gas limit fuzzing enabled for transaction replay")
	}

	// Use block-based replay to maintain original block boundaries
	if blockProvider, ok := txProvider.(interface{ GetNextBlockTxs() ([]payload.Tx, error) }); ok {
		logger.Info("Using block-based replay that maintains original block boundaries")

		for {
			txBatch, err := blockProvider.GetNextBlockTxs()
			if err != nil {
				if err.Error() == "no more blocks available - processing completed" {
					logger.Info("Block replay completed successfully")
					return nil
				}
				logger.WithError(err).Error("Failed to get next block transactions")
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if len(txBatch) == 0 {
				continue
			}

			logger.WithFields(log.Fields{
				"block_txs": len(txBatch),
			}).Info("Broadcasting block transactions")

			// Broadcast all transactions from this block sequentially to preserve acc seq order
			successCount := 0
			timeoutErrors := 0
			sequenceErrors := 0
			outOfGasErrors := 0
			otherErrors := 0
			fuzzedCount := 0
			fallbackCount := 0

			for _, tx := range txBatch {
				originalTxBytes := tx.Bytes()
				txBytes := originalTxBytes
				wasFuzzed := false

				// Apply gas fuzzing if enabled
				if gasFuzzer != nil {
					txBytes, wasFuzzed = fuzzTransactions(txBytes, gasFuzzer, accountClient, config, logger)
					if wasFuzzed {
						fuzzedCount++
					}
				}

				txHash, err := broadcastFunc(ctx, txBytes)

				// Handle broadcast result
				if err != nil {
					errMsg := err.Error()
					// Record error for fuzzed transactions
					categorizeError(errMsg, &timeoutErrors, &sequenceErrors, &outOfGasErrors, &otherErrors, logger)

					// Try fallback for fuzzed transactions that fail with insufficient fee
					if wasFuzzed && strings.Contains(errMsg, "out of gas") {
						fallbackHash, fallbackErr, fallbackSuccess := handleFallbackBroadcast(
							ctx, originalTxBytes, broadcastFunc, errMsg, logger)

						if fallbackSuccess {
							successCount++
							fallbackCount++
							broadcastTxPace.Step(1)
							txHash = fallbackHash
						} else {
							categorizeError(fallbackErr.Error(), &timeoutErrors, &sequenceErrors, &outOfGasErrors, &otherErrors, logger)
						}
					} else {
						categorizeError(errMsg, &timeoutErrors, &sequenceErrors, &outOfGasErrors, &otherErrors, logger)
					}
				} else if txHash == "" {
					otherErrors++
				} else {
					successCount++
					broadcastTxPace.Step(1)
				}

				// Placeholder if needed later
				_ = txHash
			}

			logFields := log.Fields{
				"total":             len(txBatch),
				"success":           successCount,
				"timeout_errors":    timeoutErrors,
				"sequence_errors":   sequenceErrors,
				"out_of_gas_errors": outOfGasErrors,
				"other_errors":      otherErrors,
			}

			if gasFuzzer != nil {
				logFields["fuzzed_transactions"] = fuzzedCount
				logFields["fuzz_rate"] = fmt.Sprintf("%.1f%%", float64(fuzzedCount)/float64(len(txBatch))*100)
				if fallbackCount > 0 {
					logFields["fallback_successes"] = fallbackCount
					logFields["fallback_rate"] = fmt.Sprintf("%.1f%%", float64(fallbackCount)/float64(fuzzedCount)*100)
				}
			}

			logger.WithFields(logFields).Info("Block broadcast summary")
		}
	}

	return nil
}

func getAccountNumberSequence(
	ctx context.Context,
	client chain.Client,
	accountAddress string,
) (uint64, uint64, error) {
	logger := log.WithField("fn", "getAccountNumberSequence")

	var accNum, accSeq uint64

	err := retry.Do(func() error {
		var err error
		accNum, accSeq, err = client.GetNumberSequence(accountAddress)
		if err != nil {
			logger.WithError(err).Warning("⚠️ Error while GetNumberSequence")

			return errors.Wrap(err, "querying for account number and sequence failed")
		}

		return nil
	},
		retry.Context(ctx),
		retry.Attempts(10),
		retry.Delay(100*time.Millisecond),
		retry.MaxDelay(5*time.Second),
	)
	if err != nil {
		return 0, 0, err
	}

	return accNum, accSeq, nil
}

func createAndBroadcastInitialTxs(
	ctx context.Context,
	logger log.Logger,
	signedTxPace,
	getAccountNumberSequencePace,
	broadcastTxPace pace.Pace,
	client chain.Client,
	provider payload.TxProvider,
	fromPrivateKeys []chain.Secp256k1PrivateKey,
) error {
	initialTxs := make([]payload.Tx, 0, len(fromPrivateKeys))

	// Check if this is a MixedPayloadProvider that needs special handling
	type mixedProvider interface {
		GenerateAllInitialTxs(req payload.TxRequest) ([]payload.Tx, error)
	}

	for keyIdx, fromPrivateKey := range fromPrivateKeys {
		// fetching account number and sequence should be relatively fast, let's do one by one for each key
		accNum, accSeq, err := getAccountNumberSequence(ctx, client, fromPrivateKey.AccAddress())
		if err != nil {
			err = errors.Wrap(err, "❌ Fetching initial Tx account number and sequence failed")
			return err
		}

		getAccountNumberSequencePace.Step(1)

		txReq := payload.TxRequest{
			Keys: []chain.Secp256k1PrivateKey{
				fromPrivateKey,
			},

			From: chain.Account{
				Key:      fromPrivateKey,
				Number:   accNum,
				Sequence: accSeq,
			},

			FromIdx: keyIdx,
			TxIdx:   0,
		}

		// For MixedPayloadProvider, we need to generate initial txs for ALL underlying providers
		if mixedProvider, ok := provider.(mixedProvider); ok {
			accountInitialTxs, err := mixedProvider.GenerateAllInitialTxs(txReq)
			if err != nil {
				err = errors.Wrap(err, "❌ Generating initial Txs failed")
				return err
			}
			initialTxs = append(initialTxs, accountInitialTxs...)
		} else {
			// Standard single-provider case
			initialTx, err := provider.GenerateInitialTx(txReq)
			if err != nil {
				err = errors.Wrap(err, "❌ Generating initial Tx failed")
				return err
			}

			if initialTx != nil {
				initialTxs = append(initialTxs, initialTx)
			}
		}
	}

	if len(initialTxs) == 0 {
		logger.WithFields(log.Fields{
			"num": len(initialTxs),
		}).Infoln("✅ No initial txs to broadcast.")

		return nil
	} else {
		logger.WithFields(log.Fields{
			"num": len(initialTxs),
		}).Debugln("✅ Generated initial txs to broadcast")
	}

	// we can broadcast initial txs in parallel because accounts are unique
	pool := workerpool.New(maxParallelInitialTxsBroadcasts)

	for _, initialTx := range initialTxs {
		initialTx := initialTx

		pool.Submit(func() {
			if err := retry.Do(func() error {
				defer catcher.Catch(
					catcher.RecvLog(true),
					catcher.RecvDie(1, true),
				)

				signedTx, err := provider.BuildAndSignTx(
					client,
					initialTx,
				)
				if err != nil {
					err = errors.Wrap(err, "❌ Signing initial Tx failed")
					return err
				}

				signedTxPace.Step(1)

				txHash, err := client.Broadcast(ctx, signedTx.Bytes(), true)
				if err != nil {
					err = errors.Wrapf(err, "❌ Broadcasting initial Tx failed: %s", txHash)
					return err
				}

				broadcastTxPace.Step(1)

				logger.WithFields(log.Fields{
					"txHash": txHash,
				}).Debugln("✅ Initial Tx broadcasted")

				return nil
			},
				retry.Context(ctx),
				retry.Attempts(10),
				retry.Delay(100*time.Millisecond),
				retry.MaxDelay(5*time.Second),
			); err != nil {
				logger.WithError(err).Error("❌ All attempts to broadcast initial Tx failed")
			}
		})
	}

	defer pool.StopWait()

	return nil
}

// buildBroadcastClient creates a broadcast retrying function with optional rate limiting
func buildBroadcastClient(config StressConfig) (ratelimit.BroadcastFunc, error) {
	baseClient := chain.NewClient(config.ChainID, config.NodeAddress, config.GRPCAddress)

	var rateLimiter *ratelimit.MultiLimiter
	if config.RateLimit.IsEnabled() {
		limiter, err := ratelimit.NewMultiLimiter(config.RateLimit, baseClient.TxConfig())
		if err != nil {
			return nil, errors.Wrap(err, "failed to create rate limiter")
		}

		rateLimiter = limiter
	}

	broadcastFn := func(ctx context.Context, txBytes []byte) (string, error) {
		if rateLimiter != nil {
			if err := rateLimiter.WaitForTransactionBytes(ctx, txBytes); err != nil {
				return "", errors.Wrap(err, "wait cancelled")
			}
		}

		var txHash string
		err := retry.Do(func() error {
			var err error

			txHash, err = baseClient.Broadcast(ctx, txBytes, config.AwaitTxConfirmation)
			if err != nil {
				if _, ok := chain.IsSequenceError(err); ok {
					return retry.Unrecoverable(err)
				}
				return errors.Wrap(err, "broadcasting transaction failed")
			}

			return nil
		},
			retry.Context(ctx),
			retry.Attempts(10),
			retry.DelayType(retry.BackOffDelay),
			retry.MaxDelay(5*time.Second),
		)
		if err != nil {
			return "", err
		}

		return txHash, nil
	}

	return broadcastFn, nil
}

func startMempoolReporter(ctx context.Context, client chain.Client, logger log.Logger) func() {
	reportCtx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-reportCtx.Done():
				return
			case <-ticker.C:
				reqCtx, reqCancel := context.WithTimeout(reportCtx, 5*time.Second)
				count, total, totalBytes, err := client.NumUnconfirmedTxs(reqCtx)
				reqCancel()

				if err != nil {
					logger.WithError(err).Warning("⚠️ Failed to query num_unconfirmed_txs")
					continue
				}

				logger.WithFields(log.Fields{
					"num_unconfirmed_txs": count,
					"total_txs":           total,
					"total_bytes":         totalBytes,
				}).Info("Mempool status")
			}
		}
	}()

	return cancel
}
