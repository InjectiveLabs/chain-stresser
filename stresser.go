package stresser

import (
	"context"
	"fmt"
	"runtime"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/dottedmag/parallel"
	"github.com/pkg/errors"
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

var errRetry = errors.New("retry required")

func Stress(
	ctx context.Context,
	config StressConfig,
	txProvider payload.TxProvider,
) error {
	logger := log.WithField("bench", txProvider.Name())
	client := chain.NewClient(config.ChainID, config.NodeAddress)

	startTs := time.Now()
	signedTxPace := pace.New("signed tx", 10*time.Second, newPaceReporter(logger))
	getAccountNumberSequencePace := pace.New("sequence fetched", 10*time.Second, newPaceReporter(logger))

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

			if len(config.Accounts) == 0 {
				return errors.New("empty accounts list")
			} else {
				// this ensures that the state required for benchmark is correctly initialized
				// for EVM transactions this usually deploys a smart contract.
				orPanic(createAndBroadcastInitialTx(ctx, logger, client, txProvider, config.Accounts[0]))
			}

			initialAccountSequences = make([]uint64, numOfAccounts)

			for fromIdx := 0; fromIdx < numOfAccounts; fromIdx++ {
				fromPrivateKey := config.Accounts[fromIdx]

				accNum, accSeq, err := getAccountNumberSequence(ctx, client, fromPrivateKey.AccAddress())
				if err != nil {
					err = errors.Wrap(err, "❌ Fetching account number and sequence failed")
					return err
				}

				getAccountNumberSequencePace.Step(1)
				initialAccountSequences[fromIdx] = accSeq

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
						return ctx.Err()
					case txQueue <- tx:
					}

					txRequest.From.Sequence++
				}
			}

			return nil
		})

		spawn("collect", parallel.Exit, func(ctx context.Context) error {
			defer func() {
				signedTxPace.Pause()
			}()

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

	broadcastTxPace := pace.New("sent tx", 10*time.Second, newPaceReporter(logger))
	startTs = time.Now()

	logger.Info("Broadcasting transactions 🚀")
	if err = parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("accounts", parallel.Exit, func(ctx context.Context) error {
			return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
				for i, accountTxs := range signedTxs {
					accountTxs := accountTxs
					initialSequence := initialAccountSequences[i]
					accountClient := chain.NewClient(config.ChainID, config.NodeAddress)

					spawn(fmt.Sprintf("account-%d", i), parallel.Continue, func(ctx context.Context) error {
						for txIndex := 0; txIndex < config.NumOfTransactions; {

							var txHash string

							if finalErr := retry.Do(
								func() (err error) {
									tx := accountTxs[txIndex]

									txHash, err = accountClient.Broadcast(ctx, tx, config.AwaitTxConfirmation)
									if err != nil {
										if expectedAccSeq, ok := chain.IsSequenceError(err); ok {
											logger.WithError(err).WithFields(log.Fields{
												"expectedSequence": expectedAccSeq,
											}).Warning("⚠️ Tx broadcasting failed, trying suggested sequence")

											txIndex = int(expectedAccSeq - initialSequence)
											return errRetry
										} else if chain.IsMempoolFullError(err) {
											return errRetry
										}

										err = errors.Wrap(err, "⚠️ Tx broadcasting error")
										return retry.Unrecoverable(err)
									}

									return nil
								},
								retry.UntilSucceeded(),
								retry.Delay(time.Second),
							); finalErr != nil {
								return finalErr
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

func createAndBroadcastInitialTx(
	ctx context.Context,
	logger log.Logger,
	client chain.Client,
	provider payload.TxProvider,
	fromPrivateKey chain.Secp256k1PrivateKey,
) error {
	accNum, accSeq, err := getAccountNumberSequence(ctx, client, fromPrivateKey.AccAddress())
	if err != nil {
		err = errors.Wrap(err, "❌ Fetching deployer account number and sequence failed")
		return err
	}

	initialTx, err := provider.GenerateInitialTx(payload.TxRequest{
		Keys: []chain.Secp256k1PrivateKey{
			fromPrivateKey,
		},

		From: chain.Account{
			Name:     "deployer",
			Key:      fromPrivateKey,
			Number:   accNum,
			Sequence: accSeq,
		},

		FromIdx: 0,
	})
	if err != nil {
		err = errors.Wrap(err, "❌ Generating initial Tx failed")
		return err
	}

	if initialTx == nil {
		// skip init tx
		return nil
	}

	signedTx, err := provider.BuildAndSignTx(
		client,
		initialTx,
	)
	if err != nil {
		err = errors.Wrap(err, "❌ Signing initial Tx failed")
		return err
	}

	txHash, err := client.Broadcast(ctx, signedTx.Bytes(), true)
	if err != nil {
		err = errors.Wrapf(err, "❌ Broadcasting initial Tx failed: %s", txHash)
		return err
	}

	logger.WithFields(log.Fields{
		"txHash": txHash,
	}).Infoln("✅ Initial Tx broadcasted")

	return nil
}
