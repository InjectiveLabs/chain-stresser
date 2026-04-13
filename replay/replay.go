package replay

type TxReplayConfig struct {
	CometRPC    string `yaml:"cometRPC" arg:"--comet-rpc" help:"CometBFT JSON-RPC endpoint to query for transactions in blocks"`
	StartHeight int64  `yaml:"startHeight" arg:"--start-height" help:"block height to start sniffing transactions from"`
	EndHeight   int64  `yaml:"endHeight" arg:"--end-height" help:"block height to end sniffing transactions from. 0 = endless mode (real-time sniffing)"`
	FromFile    string `yaml:"fromFile" arg:"--from-file" help:"filepath to read sniffed txs from. If omitted, will sniff txs from RPC in realtime."`
	ToFile      string `yaml:"toFile" arg:"--to-file" help:"filepath to store sniffed txs into. If omitted, will replay sniffed txs in realtime."`
}
