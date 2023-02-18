package main

import (
	cli "github.com/jawher/mow.cli"
)

// initGlobalOptions defines some global CLI options, that are useful for most parts of the app.
// Before adding option to there, consider moving it into the actual Cmd.
func initGlobalOptions(
	appLogLevel **string,
) {
	*appLogLevel = app.String(cli.StringOpt{
		Name:   "l log-level",
		Desc:   "Available levels: error, warn, info, debug.",
		EnvVar: "EXCHANGE_LOG_LEVEL",
		Value:  "info",
	})
}

func initCosmosOptions(
	c *cli.Cmd,
	cosmosChainID **string,
	cosmosGRPC **string,
	tendermintRPC **string,
	keyName **string,
) {
	*cosmosChainID = c.String(cli.StringOpt{
		Name:   "cosmos-chain-id",
		Desc:   "Specify Chain ID of the Cosmos network.",
		EnvVar: "STRESSER_COSMOS_CHAIN_ID",
		Value:  "injective-1",
	})

	*cosmosGRPC = c.String(cli.StringOpt{
		Name:   "cosmos-grpc",
		Desc:   "Cosmos GRPC querying endpoint",
		EnvVar: "STRESSER_COSMOS_GRPC",
		Value:  "tcp://localhost:9900",
	})

	*tendermintRPC = c.String(cli.StringOpt{
		Name:   "tendermint-rpc",
		Desc:   "Tendermint RPC endpoint",
		EnvVar: "STRESSER_TENDERMINT_RPC",
		Value:  "http://localhost:26657",
	})

	*keyName = c.String(cli.StringOpt{
		Name:   "K key-name",
		Desc:   "Keyring key name to use. Specify 'env' to use one loaded from STRESSER_ACCOUNT_MNEMONIC",
		EnvVar: "STRESSER_KEY_NAME",
		Value:  "user1",
	})
}
