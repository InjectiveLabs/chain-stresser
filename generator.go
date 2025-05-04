package stresser

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	tmed25519 "github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

type GeneratorEnvironment struct {
	ChainID                  string
	EthChainID               int
	EvmEnabled               bool
	NumOfValidators          int
	NumOfSentryNodes         int
	NumOfInstances           int
	NumOfAccountsPerInstance int
	OutDirectory             string
	ProdLike                 bool
}

const (
	bondDenom = "inj"

	// initialBalanceStaker to be 100K INJ = 100000 * 10^18 inj
	initialBalanceStaker = "100000000000000000000000" + bondDenom

	// initialBalanceBonded to be 10K INJ = 10000 * 10^18 inj
	initialBalanceBonded = "10000000000000000000000" + bondDenom

	// initialBalanceAccount to be 1M INJ = 1000000 * 10^18 inj
	initialBalanceAccount = "1000000000000000000000000" + bondDenom

	// minimumGasPrices to be used for realistic bench (involving x/distribition)
	minimumGasPrices = "1inj"
)

func GenerateConfigs(
	env GeneratorEnvironment,
) {
	if env.NumOfInstances <= 0 {
		panic("number of instances must be greater than 0")
	}

	if env.NumOfValidators <= 0 {
		panic("number of validators must be greater than 0")
	}

	dir := env.OutDirectory + "/chain-stresser-deploy"
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	genesis := chain.NewGenesis(&chain.GenesisConfig{
		ChainID:    env.ChainID,
		EthChainID: env.EthChainID,
		EvmEnabled: env.EvmEnabled,
		ProdLike:   env.ProdLike,
	})

	nodeIDs := make([]string, 0, env.NumOfValidators)
	for i := 0; i < env.NumOfValidators; i++ {
		nodePrivateKey := tmed25519.GenPrivKey()
		validatorPrivateKey := tmed25519.GenPrivKey()
		nodeIDs = append(nodeIDs, chain.NodeID(nodePrivateKey.PubKey()))

		stakerPublicKey, stakerPrivateKey := chain.GenerateSecp256k1Key()

		valDir := fmt.Sprintf("%s/validators/%d", dir, i)

		txIndexerKind := chain.TxIndexerKV
		if env.ProdLike {
			txIndexerKind = chain.TxIndexerDisabled
		}

		nodeConfig := &chain.NodeConfig{
			Moniker:              fmt.Sprintf("validator-%d", i),
			IP:                   net.IPv4zero,
			PrometheusPort:       chain.DefaultPorts.Prometheus,
			NodeKey:              nodePrivateKey,
			ValidatorKey:         validatorPrivateKey,
			ProdLike:             env.ProdLike,
			TxIndexer:            txIndexerKind,
			DiscardABCIResponses: env.ProdLike, // discard in prod
		}
		nodeConfig.Save(valDir)

		appConfig := &chain.AppConfig{
			MinimumGasPrices: minimumGasPrices,
			EVMEnabled:       env.EvmEnabled,
			ProdLike:         env.ProdLike,
		}
		appConfig.Save(valDir)

		genesis.AddAccount(stakerPublicKey.Address(), initialBalanceStaker)
		genesis.AddValidator(validatorPrivateKey.PubKey(), stakerPrivateKey, initialBalanceBonded)
	}
	orPanic(os.WriteFile(dir+"/validators/ids.json", bytesOrPanic(json.Marshal(nodeIDs)), 0o600))

	for i := 0; i < env.NumOfInstances; i++ {
		accounts := make([]chain.Secp256k1PrivateKey, 0, env.NumOfAccountsPerInstance)

		for j := 0; j < env.NumOfAccountsPerInstance; j++ {
			accountPublicKey, accountPrivateKey := chain.GenerateSecp256k1Key()
			accounts = append(accounts, accountPrivateKey)
			genesis.AddAccount(accountPublicKey.Address(), initialBalanceAccount)
		}

		instanceDir := fmt.Sprintf("%s/instances/%d", dir, i)
		orPanic(os.MkdirAll(instanceDir, 0o700))

		accountsJSON := bytesOrPanic(json.Marshal(accounts))
		orPanic(os.WriteFile(instanceDir+"/accounts.json", accountsJSON, 0o600))
	}

	for i := 0; i < env.NumOfValidators; i++ {
		genesis.Save(fmt.Sprintf("%s/validators/%d", dir, i))
	}

	if env.NumOfSentryNodes > 0 {
		nodeIDs = make([]string, 0, env.NumOfSentryNodes)
		for i := 0; i < env.NumOfSentryNodes; i++ {
			nodePrivateKey := tmed25519.GenPrivKey()

			nodeConfig := &chain.NodeConfig{
				Moniker:              fmt.Sprintf("sentry-node-%d", i),
				IP:                   net.IPv4zero,
				PrometheusPort:       chain.DefaultPorts.Prometheus,
				NodeKey:              nodePrivateKey,
				ProdLike:             env.ProdLike,
				TxIndexer:            chain.TxIndexerKV,
				DiscardABCIResponses: false,
			}

			appConfig := &chain.AppConfig{
				MinimumGasPrices: minimumGasPrices,
				EVMEnabled:       env.EvmEnabled,
				ProdLike:         env.ProdLike,
			}

			nodeDir := fmt.Sprintf("%s/sentry-nodes/%d", dir, i)
			nodeConfig.Save(nodeDir)
			appConfig.Save(nodeDir)
			genesis.Save(nodeDir)

			nodeIDs = append(nodeIDs, chain.NodeID(nodePrivateKey.PubKey()))
		}

		idsJSON := bytesOrPanic(json.Marshal(nodeIDs))
		orPanic(os.WriteFile(dir+"/sentry-nodes/ids.json", idsJSON, 0o600))
	}
}
