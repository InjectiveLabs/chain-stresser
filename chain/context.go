package chain

import (
	chainsdk "github.com/InjectiveLabs/sdk-go/client/chain"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	"google.golang.org/grpc"
)

func NewContext(chainID string, rpcClient rpcclient.Client, grpcClient *grpc.ClientConn) client.Context {
	clientContext, err := chainsdk.NewClientContext(
		chainID, "", nil,
	)
	orPanic(err)

	clientContext = clientContext.WithClient(rpcClient).WithGRPCClient(grpcClient)

	return clientContext
}
