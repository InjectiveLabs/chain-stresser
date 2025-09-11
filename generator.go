package stresser

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

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
	DockerImage              string
	DockerSubnet             string
	OutDirectory             string
	ProdLike                 bool
	Debug                    bool
	LocalNative              bool
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

	rootOutDir := env.OutDirectory + "/chain-stresser-deploy"
	if err := os.RemoveAll(rootOutDir); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	genesis := chain.NewGenesis(&chain.GenesisConfig{
		ChainID:    env.ChainID,
		EthChainID: env.EthChainID,
		EvmEnabled: env.EvmEnabled,
		ProdLike:   env.ProdLike,
	})

	persistentValidatorPeers := make([]string, 0, env.NumOfValidators)
	validatorNodeIDs := make([]string, 0, env.NumOfValidators)
	allNodeConfigs := make([]chain.NodeConfig, 0, env.NumOfValidators+env.NumOfSentryNodes)

	for i := 0; i < env.NumOfValidators; i++ {
		nodePrivateKey := tmed25519.GenPrivKey()
		validatorPrivateKey := tmed25519.GenPrivKey()

		stakerPublicKey, stakerPrivateKey := chain.GenerateSecp256k1Key()

		relativeNodeDir := fmt.Sprintf("validators/%d", i)
		valDir := fmt.Sprintf("%s/%s", rootOutDir, relativeNodeDir)

		txIndexerKind := chain.TxIndexerKV
		if env.ProdLike {
			txIndexerKind = chain.TxIndexerDisabled
		}

		nodeIPAddr := net.IPv4(127, 0, 0, 1)
		if !env.LocalNative {
			nodeIPAddr = ipInSubnet(env.DockerSubnet, i+1)
		}

		nodePorts := chain.DefaultPorts
		if env.LocalNative {
			nodePorts = shiftPorts(nodePorts, i)
		}

		nodeConfig := &chain.NodeConfig{
			Home:                 fmt.Sprintf("./%s", relativeNodeDir),
			Moniker:              fmt.Sprintf("validator-%d", i),
			PeerID:               chain.NodeID(nodePrivateKey.PubKey()),
			IPListen:             net.IPv4zero,
			IPAddr:               nodeIPAddr,
			Ports:                nodePorts,
			NodeKey:              nodePrivateKey,
			ValidatorKey:         validatorPrivateKey,
			ProdLike:             env.ProdLike,
			TxIndexer:            txIndexerKind,
			DiscardABCIResponses: env.ProdLike,                        // discard in prod
			PortsExposed:         env.NumOfSentryNodes == 0 && i == 0, // expose ports only for the first validator (and no sentry nodes)
		}
		nodeConfig.Save(valDir)

		allNodeConfigs = append(allNodeConfigs, *nodeConfig)
		validatorNodeIDs = append(validatorNodeIDs, nodeConfig.PeerID)
		persistentValidatorPeers = append(persistentValidatorPeers,
			fmt.Sprintf("%s@%s:%d", validatorNodeIDs[i], nodeConfig.IPAddr.String(), nodeConfig.Ports.P2P),
		)

		appConfig := &chain.AppConfig{
			MinimumGasPrices: minimumGasPrices,
			EVMEnabled:       env.EvmEnabled,
			ProdLike:         env.ProdLike,
			IPListen:         net.IPv4zero,
			Ports:            nodePorts,
			PortsExposed:     env.NumOfSentryNodes == 0 && i == 0, // expose ports only for the first validator (and no sentry nodes)
		}
		appConfig.Save(valDir)

		genesis.AddAccount(stakerPublicKey.Address(), initialBalanceStaker)
		genesis.AddValidator(validatorPrivateKey.PubKey(), stakerPrivateKey, initialBalanceBonded)
	}
	orPanic(os.WriteFile(rootOutDir+"/validators/ids.json", bytesOrPanic(json.Marshal(validatorNodeIDs)), 0o644))

	for i := 0; i < env.NumOfInstances; i++ {
		accounts := make([]chain.Secp256k1PrivateKey, 0, env.NumOfAccountsPerInstance)

		for j := 0; j < env.NumOfAccountsPerInstance; j++ {
			accountPublicKey, accountPrivateKey := chain.GenerateSecp256k1Key()
			accounts = append(accounts, accountPrivateKey)
			genesis.AddAccount(accountPublicKey.Address(), initialBalanceAccount)
		}

		instanceDir := fmt.Sprintf("%s/instances/%d", rootOutDir, i)
		orPanic(os.MkdirAll(instanceDir, 0o755))

		accountsJSON := bytesOrPanic(json.Marshal(accounts))
		orPanic(os.WriteFile(instanceDir+"/accounts.json", accountsJSON, 0o644))
	}

	for i := 0; i < env.NumOfValidators; i++ {
		genesis.Save(fmt.Sprintf("%s/validators/%d", rootOutDir, i))
	}

	if env.NumOfSentryNodes > 0 {
		sentryNodeIDs := make([]string, 0, env.NumOfSentryNodes)
		for i := 0; i < env.NumOfSentryNodes; i++ {
			nodePrivateKey := tmed25519.GenPrivKey()
			relativeNodeDir := fmt.Sprintf("sentry-nodes/%d", i)
			nodeDir := fmt.Sprintf("%s/%s", rootOutDir, relativeNodeDir)

			nodeConfig := &chain.NodeConfig{
				Home:                 fmt.Sprintf("./%s", relativeNodeDir),
				Moniker:              fmt.Sprintf("sentry-node-%d", i),
				PeerID:               chain.NodeID(nodePrivateKey.PubKey()),
				IPListen:             net.IPv4zero,
				IPAddr:               ipInSubnet(env.DockerSubnet, env.NumOfValidators+i+1),
				Ports:                chain.DefaultPorts,
				NodeKey:              nodePrivateKey,
				ProdLike:             env.ProdLike,
				TxIndexer:            chain.TxIndexerKV,
				DiscardABCIResponses: false,
				PortsExposed:         i == 0, // expose ports only for the first sentry node
				DependsOn:            makeIntRange(0, env.NumOfValidators),
				PersistentPeers:      strings.Join(persistentValidatorPeers, ","),
				PrivatePeerIds:       strings.Join(validatorNodeIDs[:env.NumOfValidators], ","),
			}
			allNodeConfigs = append(allNodeConfigs, *nodeConfig)

			appConfig := &chain.AppConfig{
				MinimumGasPrices: minimumGasPrices,
				EVMEnabled:       env.EvmEnabled,
				ProdLike:         env.ProdLike,
				IPListen:         net.IPv4zero,
				Ports:            chain.DefaultPorts,
				PortsExposed:     i == 0, // expose ports only for the first sentry node
			}

			nodeConfig.Save(nodeDir)
			appConfig.Save(nodeDir)
			genesis.Save(nodeDir)

			sentryNodeIDs = append(sentryNodeIDs, nodeConfig.PeerID)
		}

		idsJSON := bytesOrPanic(json.Marshal(sentryNodeIDs))
		orPanic(os.WriteFile(rootOutDir+"/sentry-nodes/ids.json", idsJSON, 0o644))
	}

	for i := 0; i < env.NumOfValidators; i++ {
		nodeConfig := allNodeConfigs[i]
		peerAddress := fmt.Sprintf("%s@%s:%d", validatorNodeIDs[i], nodeConfig.IPAddr.String(), nodeConfig.Ports.P2P)

		nodeConfig.PersistentPeers = strings.Join(filterStringValue(persistentValidatorPeers, peerAddress), ",")
		nodeConfig.PrivatePeerIds = strings.Join(validatorNodeIDs, ",")

		allNodeConfigs[i] = nodeConfig

		// save the node config again, for every validator
		nodeConfig.Save(filepath.Join(rootOutDir, nodeConfig.Home))
	}

	if !env.LocalNative {
		chain.GenerateDockerCompose(
			env.ChainID,
			env.DockerImage,
			env.DockerSubnet,
			allNodeConfigs,
			rootOutDir,
			env.Debug,
		)
	}
}

func filterStringValue(list []string, filter string) []string {
	newList := make([]string, 0, len(list))
	for _, v := range list {
		if v != filter {
			newList = append(newList, v)
		}
	}

	return newList
}

func ipInSubnet(subnet string, offset int) net.IP {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic(err)
	}

	// Get the first IP in subnet
	ip := ipNet.IP.To4()
	if ip == nil {
		panic("only IPv4 subnets supported")
	}

	// Add offset to last octet
	ip[3] += byte(offset + 1) // +1 because we want to start from the second IP in the subnet (not gateway)

	return ip
}

func shiftPorts(ports chain.Ports, offset int) chain.Ports {
	// digitOffset is 100 then 26657 -> 26757 -> 26857 -> 26957
	const digitOffset = 100

	return chain.Ports{
		RPC:        ports.RPC + digitOffset*offset,
		P2P:        ports.P2P + digitOffset*offset,
		API:        ports.API + digitOffset*offset,
		GRPC:       ports.GRPC + digitOffset*offset,
		GRPCWeb:    ports.GRPCWeb + digitOffset*offset,
		PProf:      ports.PProf + digitOffset*offset,
		Prometheus: ports.Prometheus + digitOffset*offset,
		EVMRPC:     ports.EVMRPC + digitOffset*offset,
		EVMWSPort:  ports.EVMWSPort + digitOffset*offset,
		ProxyApp:   ports.ProxyApp + digitOffset*offset,
	}
}

func makeIntRange(start, end int) []int {
	r := make([]int, end-start)
	for i := range r {
		r[i] = start + i
	}

	return r
}
