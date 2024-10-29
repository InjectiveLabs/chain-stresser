package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xlab/closer"
	log "github.com/xlab/suplog"

	stresser "github.com/InjectiveLabs/chain-stresser/v2"
	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/InjectiveLabs/chain-stresser/v2/payload"
)

const (
	defaultChainID       = "stressinj-1337"
	defaultMinGasPrice   = "1inj"
	defaultNumOfAccounts = 1000
	defaultNumOfTx       = 100

	defaultNumOfValidators = 1
	defaultNumOfSentries   = 0
	defaultNumOfInstances  = 1
)

func init() {
	// ignore debugging stuff by default
	log.DefaultLogger.SetLevel(log.InfoLevel)
}

func main() {
	var (
		stressCfg = stresser.StressConfig{
			ChainID:           defaultChainID,
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

	rootCmd.PersistentFlags().StringVar(&stressCfg.ChainID, "chain-id", defaultChainID, "Expected ID of the chain.")
	rootCmd.PersistentFlags().StringVar(&stressCfg.MinGasPrice, "min-gas-price", defaultMinGasPrice, "Minimum gas price to pay for each transaction.")
	rootCmd.PersistentFlags().StringVar(&stressCfg.NodeAddress, "node-addr", "localhost:26657", "Address of a injectived node RPC to connect to.")
	rootCmd.PersistentFlags().BoolVar(&stressCfg.AwaitTxConfirmation, "await", false, "Await for transaction to be included in a block.")
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
			stresser.GenerateConfigs(genEnv)
			return nil
		},
	}

	generateCmd.Flags().StringVar(&genEnv.ChainID, "chain-id", defaultChainID, "ID of the chain to generate.")
	generateCmd.Flags().BoolVar(&genEnv.EvmEnabled, "evm", false, "Enabled EVM support. Generates genesis with EVM state.")
	generateCmd.Flags().IntVar(&genEnv.NumOfValidators, "validators", defaultNumOfValidators, "Number of validators to generate config for.")
	generateCmd.Flags().IntVar(&genEnv.NumOfSentryNodes, "sentries", defaultNumOfSentries, "Number of sentry nodes to generate config for.")
	generateCmd.Flags().IntVar(&genEnv.NumOfInstances, "instances", defaultNumOfInstances, "The maximum number of parallel chain-stresser instances to be prepared for.")
	generateCmd.Flags().IntVar(&genEnv.NumOfAccountsPerInstance, "accounts-num", defaultNumOfAccounts, "Number of funded accounts to generate for each instance.")
	generateCmd.Flags().StringVar(&genEnv.OutDirectory, "out", strOrPanic(os.Getwd()), "Path to the directory where generated files are stored.")
	rootCmd.AddCommand(generateCmd)

	txBankSendCmd := &cobra.Command{
		Use:   "tx-bank-send",
		Short: "Run stresstest with x/bank.MsgSend transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	txEthSendCmd := &cobra.Command{
		Use:   "tx-eth-send",
		Short: "Run stresstest with eth value send transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			sendAmount := "1" + chain.DefaultBondDenom
			ethSendProvider, err := payload.NewEthSendProvider(stressCfg.ChainID, stressCfg.MinGasPrice, sendAmount)
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
			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethCallProvider, err := payload.NewEthCallProvider(stressCfg.ChainID, stressCfg.MinGasPrice)
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
			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethDeployProvider, err := payload.NewEthDeployProvider(stressCfg.ChainID, stressCfg.MinGasPrice)
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
			orPanic(readAccounts(&stressCfg, accountFile, numOfAccounts))

			ethCallProvider, err := payload.NewEthInternalCallProvider(
				stressCfg.ChainID,
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
	txEthInternalCallCmd.Flags().Uint64Var(&ethInternalCallIterations, "iterations", 10000, "Number of internal call iterations to run for each external tx")
	rootCmd.AddCommand(txEthInternalCallCmd)

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
