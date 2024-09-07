package chain

import (
	"encoding/hex"
	"net"
	"os"
	"strconv"

	bftconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
)

type NodeConfig struct {
	Name           string
	IP             net.IP
	PrometheusPort int
	NodeKey        crypto.PrivKey
	ValidatorKey   crypto.PrivKey
}

func (vc *NodeConfig) Save(homeDir string) {
	orPanic(os.MkdirAll(homeDir+"/config", 0o700))

	orPanic((&p2p.NodeKey{
		PrivKey: vc.NodeKey,
	}).SaveAs(homeDir + "/config/node_key.json"))

	if vc.ValidatorKey != nil {
		orPanic(os.MkdirAll(homeDir+"/data", 0o700))

		privval.NewFilePV(
			vc.ValidatorKey,
			homeDir+"/config/priv_validator_key.json",
			homeDir+"/data/priv_validator_state.json",
		).Save()
	}

	cfg := bftconfig.DefaultConfig()
	cfg.Moniker = vc.Name
	// set addr_book_strict to false so nodes connecting from non-routable hosts are added to address book
	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	cfg.P2P.MaxNumOutboundPeers = 100
	cfg.P2P.MaxNumInboundPeers = 100
	cfg.RPC.MaxSubscriptionClients = 10000
	cfg.RPC.MaxOpenConnections = 10000
	cfg.RPC.GRPCMaxOpenConnections = 10000
	cfg.RPC.MaxSubscriptionsPerClient = 10000
	cfg.Mempool.Size = 50000
	cfg.Mempool.MaxTxsBytes = 5368709120
	cfg.Instrumentation.Prometheus = true
	cfg.Instrumentation.PrometheusListenAddr = net.JoinHostPort(vc.IP.String(), strconv.Itoa(vc.PrometheusPort))
	bftconfig.WriteConfigFile(homeDir+"/config/config.toml", cfg)
}

// NodeID computes node ID from node public key
func NodeID(pubKey crypto.PubKey) string {
	return hex.EncodeToString(pubKey.Address())
}
