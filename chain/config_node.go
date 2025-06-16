package chain

import (
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"strconv"
	"text/template"

	bftconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"

	_ "embed"
)

type TxIndexerKind string

const (
	TxIndexerKV       TxIndexerKind = "kv"
	TxIndexerDisabled TxIndexerKind = "null"
)

type NodeConfig struct {
	Home                 string
	Moniker              string
	PeerID               string
	IPListen             net.IP
	IPAddr               net.IP
	Ports                Ports
	PersistentPeers      string
	PrivatePeerIds       string
	NodeKey              crypto.PrivKey
	ValidatorKey         crypto.PrivKey
	TxIndexer            TxIndexerKind
	DiscardABCIResponses bool
	ProdLike             bool
	PortsExposed         bool

	// DependsOn only used for docker-compose
	// Is a list of validator indices that this node depends on.
	DependsOn []int
}

func (nodeConfig *NodeConfig) Save(homeDir string) {
	orPanic(os.MkdirAll(homeDir+"/config", 0o700))

	orPanic((&p2p.NodeKey{
		PrivKey: nodeConfig.NodeKey,
	}).SaveAs(homeDir + "/config/node_key.json"))

	if nodeConfig.ValidatorKey != nil {
		orPanic(os.MkdirAll(homeDir+"/data", 0o700))

		privval.NewFilePV(
			nodeConfig.ValidatorKey,
			homeDir+"/config/priv_validator_key.json",
			homeDir+"/data/priv_validator_state.json",
		).Save()
	}

	cfg := bftconfig.DefaultConfig()

	if nodeConfig.ProdLike {
		buf := new(bytes.Buffer)
		tpl := template.Must(template.New("config_prod").Parse(string(configProdTplTOML)))
		orPanic(tpl.Execute(buf, nodeConfig))
		orPanic(os.WriteFile(homeDir+"/config/config.toml", buf.Bytes(), 0o600))
		return
	}

	// set addr_book_strict to false so nodes connecting from non-routable hosts are added to address book
	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	cfg.P2P.MaxNumOutboundPeers = 100
	cfg.P2P.MaxNumInboundPeers = 100
	cfg.P2P.PersistentPeers = nodeConfig.PersistentPeers
	cfg.P2P.PrivatePeerIDs = nodeConfig.PrivatePeerIds
	cfg.RPC.MaxSubscriptionClients = 10000
	cfg.RPC.MaxOpenConnections = 10000
	cfg.RPC.MaxSubscriptionsPerClient = 10000
	cfg.Mempool.Size = 50000
  cfg.Mempool.MaxTxsBytes = 671088640

	if nodeConfig.PortsExposed {
		cfg.P2P.ListenAddress = "tcp://" + net.JoinHostPort(
			nodeConfig.IPListen.String(),
			strconv.Itoa(nodeConfig.Ports.P2P),
		)
		cfg.RPC.ListenAddress = "tcp://" + net.JoinHostPort(
			nodeConfig.IPListen.String(),
			strconv.Itoa(nodeConfig.Ports.RPC),
		)

		cfg.Instrumentation.Prometheus = true
		cfg.Instrumentation.PrometheusListenAddr = net.JoinHostPort(
			nodeConfig.IPListen.String(),
			strconv.Itoa(nodeConfig.Ports.Prometheus),
		)
	}

	cfg.Moniker = nodeConfig.Moniker
	cfg.TxIndex.Indexer = string(TxIndexerKV)
	cfg.Storage.DiscardABCIResponses = false

	bftconfig.WriteConfigFile(homeDir+"/config/config.toml", cfg)
}

// NodeID computes node ID from node public key
func NodeID(pubKey crypto.PubKey) string {
	return hex.EncodeToString(pubKey.Address())
}

var (
	//go:embed templates/config.prod.toml.tpl
	configProdTplTOML []byte
)
