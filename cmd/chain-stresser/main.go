package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xlab/closer"
	"github.com/xlab/pace"
	log "github.com/xlab/suplog"

	stresser "github.com/InjectiveLabs/chain-stresser/v2"
	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/InjectiveLabs/chain-stresser/v2/payload"
	"github.com/InjectiveLabs/chain-stresser/v2/state"
)

const (
	defaultChainID       = "stressinj-1337"
	defaultEthChainID    = 1337
	defaultMinGasPrice   = "1inj"
	defaultNumOfAccounts = 1000
	defaultNumOfTx       = 100

	defaultNumOfValidators = 1
	defaultNumOfSentries   = 0
	defaultNumOfInstances  = 1

	defaultInjectiveDockerImage = "injectivelabs/injective-core"
	defaultDockerSubnet         = "172.127.0.0/24"

	// TODO: update to latest version when it's released
	latestInjectiveCoreTag = "v1.16.0-beta.3"
)

var (
	verboseOutput = false
)

func init() {
	// ignore debugging stuff by default
	log.DefaultLogger.SetLevel(log.InfoLevel)
}

func main() {
	var (
		stressCfg = stresser.StressConfig{
			EthChainID:        defaultEthChainID,
			MinGasPrice:       defaultMinGasPrice,
			NumOfTransactions: defaultNumOfTx,
		}

		accountFile   string = "accounts.json"
		numOfAccounts int    = defaultNumOfAccounts
	)

	defer closer.Close()

	rootCtx, cancelFn := context.WithCancel(context.Background())
	closer.Bind(cancelFn)

	rootCmd := &cobra.Command{
		Use: "chain-stresser",

		Hidden:        true,
		SilenceErrors: true,
		SilenceUsage:  false,
		Long:          bannerStr,

		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	rootCmd.PersistentFlags().StringVar(&stressCfg.ChainID, "chain-id", defaultChainID, "Expected Cosmos chain ID of the chain to connect to.")
	rootCmd.PersistentFlags().Int64Var(&stressCfg.EthChainID, "eth-chain-id", defaultEthChainID, "Expected EIP-155 chain ID of the EVM.")
	rootCmd.PersistentFlags().StringVar(&stressCfg.MinGasPrice, "min-gas-price", defaultMinGasPrice, "Minimum gas price to pay for each transaction.")
	rootCmd.PersistentFlags().StringVar(&stressCfg.NodeAddress, "node-addr", "127.0.0.1:26657", "Address of a injectived node RPC to connect to.")
	rootCmd.PersistentFlags().StringVar(&stressCfg.GRPCAddress, "grpc-addr", "127.0.0.1:9900", "Address of a injectived node GRPC to connect to.")
	rootCmd.PersistentFlags().BoolVar(&stressCfg.AwaitTxConfirmation, "await", true, "Await for transaction to be included in a block.")
	rootCmd.PersistentFlags().BoolVar(&verboseOutput, "verbose", false, "Verbosely output debugging information.")
	rootCmd.PersistentFlags().StringVar(&accountFile, "accounts", "accounts.json", "Path to a JSON file containing private keys of accounts to use for stress testing.")
	rootCmd.PersistentFlags().IntVar(&numOfAccounts, "accounts-num", defaultNumOfAccounts, "Number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file.")
	rootCmd.PersistentFlags().IntVar(&stressCfg.NumOfTransactions, "transactions", defaultNumOfTx, "Number of transactions to allocate for each account.")
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	var genEnv stresser.GeneratorEnvironment

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates all the config files required to start injectived cluster with state for stress testing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			if genEnv.LocalNative {
				if genEnv.NumOfValidators > 4 {
					return errors.New("number of validators must be less than or equal to 4 when running natively")
				}
				if genEnv.NumOfSentryNodes > 0 {
					return errors.New("number of sentry nodes must be 0 when running natively")
				}
			}

			stresser.GenerateConfigs(genEnv)
			return nil
		},
	}

	generateCmd.Flags().StringVar(&genEnv.ChainID, "chain-id", defaultChainID, "Cosmos chain ID of the chain to generate.")
	generateCmd.Flags().IntVar(&genEnv.EthChainID, "eth-chain-id", defaultEthChainID, "EIP-155 chain ID of the EVM (can be different from the Cosmos chain-id).")
	generateCmd.Flags().BoolVar(&genEnv.EvmEnabled, "evm", true, "Enabled EVM support. Generates genesis with EVM state.")
	generateCmd.Flags().BoolVar(&genEnv.ProdLike, "prod", false, "Generate config for prod-like chain (app/bft configs will be close to mainnet versions).")
	generateCmd.Flags().BoolVar(&genEnv.Debug, "debug", false, "Additional debug output when running in Docker Compose mode.")
	generateCmd.Flags().StringVar(&genEnv.DockerImage, "docker-image", defaultInjectiveDockerImage+":"+latestInjectiveCoreTag, "Docker image to use for the local network via docker-compose.")
	generateCmd.Flags().StringVar(&genEnv.DockerSubnet, "docker-subnet", defaultDockerSubnet, "Docker subnet to use for the local network via docker-compose.")
	generateCmd.Flags().IntVar(&genEnv.NumOfValidators, "validators", defaultNumOfValidators, "Number of validators to generate config for.")
	generateCmd.Flags().IntVar(&genEnv.NumOfSentryNodes, "sentries", defaultNumOfSentries, "Number of sentry nodes to generate config for.")
	generateCmd.Flags().IntVar(&genEnv.NumOfInstances, "instances", defaultNumOfInstances, "The maximum number of parallel chain-stresser instances to be prepared for.")
	generateCmd.Flags().BoolVar(&genEnv.LocalNative, "native", false, "Generate config compatible with running multiple binaries natively on the host machine (no docker-compose).")
	generateCmd.Flags().IntVar(&genEnv.NumOfAccountsPerInstance, "accounts-num", defaultNumOfAccounts, "Number of funded accounts to generate for each instance.")
	generateCmd.Flags().StringVar(&genEnv.OutDirectory, "out", strOrPanic(os.Getwd()), "Path to the directory where generated files are stored.")
	rootCmd.AddCommand(generateCmd)

	txBankSendCmd := &cobra.Command{
		Use:   "tx-bank-send",
		Short: "Run stresstest with x/bank.MsgSend transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			sendAmount := "1" + chain.DefaultBondDenom
			bankSendProvider, err := payload.NewBankSendProvider(stressCfg.MinGasPrice, sendAmount)
			if err != nil {
				return errors.Wrap(err, "failed to initate bank send stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, bankSendProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txBankSendCmd)

	var (
		multiSendNumTargets int
	)

	txBankMultiSendCmd := &cobra.Command{
		Use:   "tx-bank-send-many",
		Short: "Run stresstest with x/bank.MsgMultiSend transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			sendAmount := "1" + chain.DefaultBondDenom
			bankMultiSendProvider, err := payload.NewBankMultiSendProvider(stressCfg.MinGasPrice, sendAmount, multiSendNumTargets)
			if err != nil {
				return errors.Wrap(err, "failed to initate bank multi send stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, bankMultiSendProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	txBankMultiSendCmd.Flags().IntVar(&multiSendNumTargets, "targets", 50, "Number of targets to send the funds to.")
	rootCmd.AddCommand(txBankMultiSendCmd)

	txEthSendCmd := &cobra.Command{
		Use:   "tx-eth-send",
		Short: "Run stresstest with eth value send transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			sendAmount := "1" + chain.DefaultBondDenom
			ethSendProvider, err := payload.NewEthSendProvider(big.NewInt(stressCfg.EthChainID), stressCfg.MinGasPrice, sendAmount)
			if err != nil {
				return errors.Wrap(err, "failed to initate eth value send stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, ethSendProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txEthSendCmd)

	txEthCallCmd := &cobra.Command{
		Use:   "tx-eth-call",
		Short: "Run stresstest with eth contract call transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethCallProvider, err := payload.NewEthCallProvider(big.NewInt(stressCfg.EthChainID), stressCfg.MinGasPrice)
			if err != nil {
				return errors.Wrap(err, "failed to initate eth contract call stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, ethCallProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txEthCallCmd)

	txEthDeployCmd := &cobra.Command{
		Use:   "tx-eth-deploy",
		Short: "Run stresstest with eth contract deploy transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethDeployProvider, err := payload.NewEthDeployProvider(big.NewInt(stressCfg.EthChainID), stressCfg.MinGasPrice)
			if err != nil {
				return errors.Wrap(err, "failed to initate eth contract deploy stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, ethDeployProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txEthDeployCmd)

	var ethInternalCallIterations uint64
	txEthInternalCallCmd := &cobra.Command{
		Use:   "tx-eth-internal-call",
		Short: "Run stresstest with eth contract internal call transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethCallProvider, err := payload.NewEthInternalCallProvider(
				big.NewInt(stressCfg.EthChainID),
				stressCfg.MinGasPrice,
				ethInternalCallIterations,
			)
			if err != nil {
				return errors.Wrap(err, "failed to initate eth contract internal call stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, ethCallProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	txEthInternalCallCmd.Flags().Uint64Var(&ethInternalCallIterations, "iterations", 10, "Number of internal call iterations to run for each external tx")
	rootCmd.AddCommand(txEthInternalCallCmd)

	var (
		ethRPCURL             string
		entrypointAddress     string
		beneficiaryAddress    string
		accountFactoryAddress string
		counterContractAddr   string
	)

	txEthUserOpCmd := &cobra.Command{
		Use:   "tx-eth-userop",
		Short: "Run stresstest with eth contract UserOp transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			userOpsSignedPace := pace.New("userops signed", 1*time.Minute, stresser.NewPaceReporter(log.DefaultLogger))

			ethUserOpProvider, err := payload.NewEthUserOpProvider(
				ethRPCURL,
				big.NewInt(stressCfg.EthChainID),
				stressCfg.MinGasPrice,
				userOpsSignedPace,
				ethcmn.HexToAddress(entrypointAddress),
				ethcmn.HexToAddress(beneficiaryAddress),
				ethcmn.HexToAddress(accountFactoryAddress),
				ethcmn.HexToAddress(counterContractAddr),
			)
			if err != nil {
				return errors.Wrap(err, "failed to initiate eth UserOp stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, ethUserOpProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}

	txEthUserOpCmd.Flags().StringVar(&ethRPCURL, "eth-rpc-url", "http://127.0.0.1:8545", "Ethereum RPC URL")
	txEthUserOpCmd.Flags().StringVar(&entrypointAddress, "entrypoint-address", "0x586AaA4d77955b36784cADf6D9D617b952d45DA1", "EntryPoint contract address")
	txEthUserOpCmd.Flags().StringVar(&beneficiaryAddress, "beneficiary-address", "0x0000000000000000000000000000000000000000", "Beneficiary address for UserOp fees")
	txEthUserOpCmd.Flags().StringVar(&accountFactoryAddress, "factory-address", "0x0B3809304F2bAad3E0d0810B98Cc7e505C06ce89", "Account Factory contract address")
	txEthUserOpCmd.Flags().StringVar(&counterContractAddr, "counter-address", "0x590d9D4654FC262BFE72d115355db2aEb7DB902f", "Counter contract address")
	rootCmd.AddCommand(txEthUserOpCmd)

	var spotMarketIDs []string
	var derivativeMarketIDs []string

	txExchangeBatchOrdersCmd := &cobra.Command{
		Use:   "tx-exchange-batch-orders",
		Short: "Run stresstest with x/exchange.MsgBatchUpdateOrders transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			exchangeBatchOrdersProvider, err := payload.NewExchangeBatchOrdersProvider(stressCfg.MinGasPrice, spotMarketIDs, derivativeMarketIDs)
			if err != nil {
				return errors.Wrap(err, "failed to initate exchange batch orders stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, exchangeBatchOrdersProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	txExchangeBatchOrdersCmd.Flags().StringSliceVar(&spotMarketIDs, "spot-market-ids", []string{"0x1422a13427d5eabd4d8de7907c8340f7e58cb15553a9fd4ad5c90406561886f9"}, "Comma-separated list of spot market IDs to update.")
	txExchangeBatchOrdersCmd.Flags().StringSliceVar(&derivativeMarketIDs, "derivative-market-ids", []string{"0x1422a13427d5eabd4d8de7907c8340f7e58cb15553a9fd4ad5c90406561886f9"}, "Comma-separated list of derivative market IDs to update.")
	rootCmd.AddCommand(txExchangeBatchOrdersCmd)

	txWasmStoreCodeCmd := &cobra.Command{
		Use:   "tx-wasm-store-code",
		Short: "Run stresstest with x/wasm.MsgStoreCode transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			wasmStoreCodeProvider, err := payload.NewWasmStoreCodeProvider(stressCfg.MinGasPrice)
			if err != nil {
				return errors.Wrap(err, "failed to initiate wasm code store stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, wasmStoreCodeProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txWasmStoreCodeCmd)

	txWasmInitContractCmd := &cobra.Command{
		Use:   "tx-wasm-init-contract",
		Short: "Run stresstest with x/wasm.MsgInstantiateContract transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			queryClient := chain.NewClient(
				stressCfg.ChainID,
				stressCfg.NodeAddress,
				stressCfg.GRPCAddress,
			)

			wasmInitContractProvider, err := payload.NewWasmInitContractProvider(
				queryClient,
				stressCfg.MinGasPrice,
			)
			if err != nil {
				return errors.Wrap(err, "failed to initiate wasm init contract stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, wasmInitContractProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txWasmInitContractCmd)

	txWasmExecContractCmd := &cobra.Command{
		Use:   "tx-wasm-exec-contract",
		Short: "Run stresstest with x/wasm.MsgExecuteContract transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			queryClient := chain.NewClient(
				stressCfg.ChainID,
				stressCfg.NodeAddress,
				stressCfg.GRPCAddress,
			)

			wasmExecContractProvider, err := payload.NewWasmExecContractProvider(
				queryClient,
				stressCfg.MinGasPrice,
			)
			if err != nil {
				return errors.Wrap(err, "failed to initiate wasm exec stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, wasmExecContractProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	rootCmd.AddCommand(txWasmExecContractCmd)

	var (
		replayFilePath     string
		snifferEnabled     bool
		snifferRPC         string
		snifferStartHeight int
		snifferEndHeight   int
		snifferAppend      bool
	)

	txStateReplayCmd := &cobra.Command{
		Use:   "tx-state-replay",
		Short: "Run stresstest with state replay transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verboseOutput {
				log.DefaultLogger.SetLevel(log.DebugLevel)
			}

			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			queryClient := chain.NewClient(
				stressCfg.ChainID,
				stressCfg.NodeAddress,
				stressCfg.GRPCAddress,
			)
			if snifferEnabled {
				sharedBuf := &bytes.Buffer{}
				sniffer, err := state.NewSniffer(
					&state.StateConfig{
						CometRPC:    snifferRPC,
						StartHeight: snifferStartHeight,
						EndHeight:   snifferEndHeight,
						FilePath:    replayFilePath,
						Append:      snifferAppend,
					},
					sharedBuf,
				)
				if err != nil {
					return errors.Wrap(err, "failed to initiate state sniffer")
				}

				snifferErrCh := make(chan error, 1)

				if err := sniffer.Start(); err != nil {
					snifferErrCh <- err
				} else {
					close(snifferErrCh)
				}

				log.Info("Waiting for state sniffer to complete...")

				select {
				case err := <-snifferErrCh:
					if err != nil {
						sniffer.Close()
						return errors.Wrap(err, "state sniffer failed")
					}
				case <-sniffer.Done():
					log.Info("State sniffer completed successfully, starting state replay stress test...")
				}

				defer sniffer.Close()
			}

			stateReplayProvider, err := payload.NewStateReplayStressProvider(
				queryClient,
				replayFilePath,
			)
			if err != nil {
				return errors.Wrap(err, "failed to initiate state replay stress provider")
			}

			if err := stresser.Stress(rootCtx, stressCfg, stateReplayProvider); err != nil {
				log.Errorf("❌ benchmark failed:\n\n%s", err)
				os.Exit(-1)
			}

			return nil
		},
	}
	txStateReplayCmd.Flags().StringVar(&replayFilePath, "replay-file-path", "/tmp/chain-stresser/txns", "Path to the file containing the state replay transactions.")
	txStateReplayCmd.Flags().BoolVar(&snifferEnabled, "sniffer-enabled", true, "Whether to enable the state sniffer.")
	txStateReplayCmd.Flags().StringVar(&snifferRPC, "sniffer-rpc", "http://127.0.0.1:26657", "RPC endpoint to use for the state sniffer.")
	txStateReplayCmd.Flags().IntVar(&snifferStartHeight, "sniffer-start-height", 0, "Start height for the state sniffer.")
	txStateReplayCmd.Flags().IntVar(&snifferEndHeight, "sniffer-end-height", 0, "End height for the state sniffer.")
	txStateReplayCmd.Flags().BoolVar(&snifferAppend, "sniffer-append", false, "Whether to append to the replay file.")
	txStateReplayCmd.MarkFlagsRequiredTogether("sniffer-enabled", "sniffer-rpc", "sniffer-start-height", "sniffer-end-height", "replay-file-path")

	rootCmd.AddCommand(txStateReplayCmd)

	orPanic(rootCmd.Execute())
}

const bannerStr = `
┏┓┓   •    ┏┓       
┃ ┣┓┏┓┓┏┓  ┗┓╋┏┓┏┓┏┏
┗┛┛┗┗┻┗┛┗  ┗┛┗┛ ┗ ┛┛

Ultimate benchmarking tool for Injective Chain 🔥
`

func readAccounts(
	cfg *stresser.StressConfig,
	accountFile string,
	numOfAccounts int,
) error {
	if numOfAccounts <= 0 {
		return errors.New("number of accounts must be greater than 0")
	}

	keysRaw, err := os.ReadFile(accountFile)
	if err != nil {
		return errors.Wrap(err, "reading account file failed")
	} else if err := json.Unmarshal(keysRaw, &cfg.Accounts); err != nil {
		return errors.Wrap(err, "parsing account file failed")
	} else if numOfAccounts > len(cfg.Accounts) {
		return errors.New("number of accounts is greater than the number of provided private keys")
	}

	cfg.Accounts = cfg.Accounts[:numOfAccounts]
	return nil
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func strOrPanic(out string, err error) string {
	if err != nil {
		panic(err)
	}

	return out
}
