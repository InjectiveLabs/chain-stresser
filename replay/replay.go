package replay

type TxReplayConfig struct {
	CometRPC    string `yaml:"cometRPC" arg:"--comet-rpc" help:"CometBFT JSON-RPC endpoint to query for transactions in blocks"`
	StartHeight int64  `yaml:"startHeight" arg:"--start-height" help:"block height to start sniffing transactions from"`
	EndHeight   int64  `yaml:"endHeight" arg:"--end-height" help:"block height to end sniffing transactions from. 0 = endless mode (real-time sniffing)"`
}
