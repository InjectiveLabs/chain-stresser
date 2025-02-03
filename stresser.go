package stresser

import (
	"context"
	"fmt"
	"runtime"
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
)

// StressConfig is the config for stress runner
type StressConfig struct {
	//  ChainID of the chain to connect to
	ChainID string

	// MinGasPrice to use for sending transactions
	MinGasPrice string

	// RPC address of the node to connect to
	NodeAddress string

	// Account privkeys to use for sending transactions
	Accounts []chain.Secp256k1PrivateKey

	// NumOfTransactions to send per account
	NumOfTransactions int

	// AwaitTxConfirmation to wait for transaction to be included in a block
	AwaitTxConfirmation bool
}

// maxParallelInitialTxsBroadcasts is the maximum number of initial txs to broadcast in parallel.
// this is internal to the stresser and not configurable for now.
const maxParallelInitialTxsBroadcasts = 8

func Stress(
	ctx context.Context,
	config StressConfig,
	txProvider payload.TxProvider,
) error {
	logger := log.WithField("bench", txProvider.Name())
	client := chain.NewClient(config.ChainID, config.NodeAddress)

	startTs := time.Now()
	signedTxPace := pace.New("signed tx", 10*time.Second, NewPaceReporter(logger))
	getAccountNumberSequencePace := pace.New("sequence fetched", 10*time.Second, NewPaceReporter(logger))
	broadcastTxPace := pace.New("sent tx", 10*time.Second, NewPaceReporter(logger))

	numOfAccounts := len(config.Accounts)
	logger.WithFields(log.Fields{
		"num": numOfAccounts * config.NumOfTransactions,
	}).Info("Preparing signed transactions. Please wait ⏳")

	var signedTxs [][][]byte
	var initialAccountSequences []uint64

	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		txQueue := make(chan payload.Tx)
		txSignedQueue := make(chan payload.Tx)

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
			workpool := workerpool.New(runtime.NumCPU())

			for fromIdx := 0; fromIdx < numOfAccounts; fromIdx++ {
				fromPrivateKey := config.Accounts[fromIdx]
				accAddress := fromPrivateKey.AccAddress()

				workpool.Submit(func() {
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

			workpool.StopWait()
			return nil
		})

		spawn("collect", parallel.Exit, func(ctx context.Context) error {
			defer func() {
				signedTxPace.Pause()
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
	if err = parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("accounts", parallel.Exit, func(ctx context.Context) error {
			return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
				for accountIdx, accountTxs := range signedTxs {
					accountTxs := accountTxs
					accountIdx := accountIdx

					initialSequence := initialAccountSequences[accountIdx]
					accountClient := chain.NewClient(config.ChainID, config.NodeAddress)

					spawn(fmt.Sprintf("account-%d", accountIdx), parallel.Continue, func(ctx context.Context) error {
						defer catcher.Catch(
							catcher.RecvLog(true),
							catcher.RecvDie(1, true),
						)

						for txIndex := 0; txIndex < config.NumOfTransactions; {
							tx := accountTxs[txIndex]

							txHash, err := accountClient.Broadcast(ctx, tx, config.AwaitTxConfirmation)
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

	for keyIdx, fromPrivateKey := range fromPrivateKeys {
		// fetching account number and sequence should be relatively fast, let's do one by one for each key
		accNum, accSeq, err := getAccountNumberSequence(ctx, client, fromPrivateKey.AccAddress())
		if err != nil {
			err = errors.Wrap(err, "❌ Fetching initial Tx account number and sequence failed")
			return err
		}

		getAccountNumberSequencePace.Step(1)

		// generating initial tx for each key
		initialTx, err := provider.GenerateInitialTx(payload.TxRequest{
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
		})
		if err != nil {
			err = errors.Wrap(err, "❌ Generating initial Tx failed")
			return err
		}

		if initialTx != nil {
			initialTxs = append(initialTxs, initialTx)
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
				retry.Attempts(5),
				retry.MaxDelay(5*time.Second),
			); err != nil {
				logger.WithError(err).Error("❌ All attempts to broadcast initial Tx failed")
			}
		})
	}

	defer pool.StopWait()

	return nil
}
