package state

type StateConfig struct {
	SnifferEnabled bool   `yaml:"snifferEnabled" arg:"--sniffer-enabled"`
	CometRPC       string `yaml:"cometRPC" arg:"--comet-rpc" help:"CometBFT JSON-RPC endpoint to query for transactions in blocks"`
	StartHeight    uint64 `yaml:"startHeight" arg:"--start-height" help:"block height to start sniffing transactions from"`
	EndHeight      uint64 `yaml:"endHeight" arg:"--end-height" help:"block height to end sniffing transactions from, optional. 0 = do not stop"`
	FilePath       string `yaml:"filePath" arg:"--file" help:"path to file where to flush sniffed transactions into. Optional, then sniffed txns can only be replayed in realtime by the broadcaster."`
	Append         bool   `yaml:"append" arg:"--append" help:"whether to append or overwrite file with new sniffed txns. Default false."`
}
