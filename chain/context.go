package chain

import (
	chainsdk "github.com/InjectiveLabs/sdk-go/client/chain"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
)

func NewContext(chainID string, rpcClient rpcclient.Client) client.Context {
	clientContext, err := chainsdk.NewClientContext(
		chainID, "", nil,
	)
	orPanic(err)

	clientContext = clientContext.WithClient(rpcClient)

	return clientContext
}
