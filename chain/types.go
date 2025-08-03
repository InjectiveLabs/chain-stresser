package chain

import (
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

type Ports struct {
	RPC        int `json:"rpc"`
	P2P        int `json:"p2p"`
	API        int `json:"api"`
	GRPC       int `json:"grpc"`
	GRPCWeb    int `json:"grpcWeb"`
	PProf      int `json:"pprof"`
	Prometheus int `json:"prometheus"`
	EVMRPC     int `json:"evmRpc"`
	EVMWSPort  int `json:"evmWsPort"`
	ProxyApp   int `json:"proxyApp"`
}

type Account struct {
	Name     string
	Key      Secp256k1PrivateKey
	Number   uint64
	Sequence uint64
}

func (a Account) String() string {
	return fmt.Sprintf("%s@%s", a.Name, a.Key.AccAddress())
}

func (a Account) EthAddress() ethcmn.Address {
	return ethcmn.BytesToAddress(a.Key.PubKey().Address().Bytes())
}
