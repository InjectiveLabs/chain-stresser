package chain

import (
	ctypes "github.com/InjectiveLabs/sdk-go/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	setAccountPrefixes("inj")
}

// DefaultPorts are the default ports the node listens on
var DefaultPorts = Ports{
	RPC:        26657,
	P2P:        26656,
	GRPC:       9090,
	GRPCWeb:    9091,
	PProf:      6060,
	Prometheus: 26660,

	// TODO: EVM RPC
}

func setAccountPrefixes(accountAddressPrefix string) {
	// Set prefixes
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)

	ctypes.SetBech32Prefixes(config)
	ctypes.SetBip44CoinType(config)
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func bytesOrPanic(out []byte, err error) []byte {
	if err != nil {
		panic(err)
	}

	return out
}
