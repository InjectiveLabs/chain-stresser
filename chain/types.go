package chain

import (
	"fmt"
)

type Ports struct {
	RPC        int `json:"rpc"`
	P2P        int `json:"p2p"`
	GRPC       int `json:"grpc"`
	GRPCWeb    int `json:"grpcWeb"`
	PProf      int `json:"pprof"`
	Prometheus int `json:"prometheus"`

	// TODO: EVM RPC port
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
