package aa

import (
	"context"
	"math/big"
	"time"

	ethclient "github.com/InjectiveLabs/chain-stresser/v2/eth/client"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"

	contract_entrypoint "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/IEntryPoint"
)

type Entrypoint interface {
	HandleOps(ctx context.Context, userOps []PackedUserOperation) (ethcmn.Hash, error)
	HandleOpsCallData(userOps []PackedUserOperation, beneficiaryAddress ethcmn.Address) ([]byte, error)
}

type entrypoint struct {
	ethClient          ethclient.EthClient
	rpcClient          *ethrpc.Client
	entrypointAddress  ethcmn.Address
	beneficiaryAddress ethcmn.Address
	chainID            *big.Int

	entrypointContract *contract_entrypoint.IEntryPoint
	entrypointABI      *abi.ABI
}

func NewEntrypoint(
	ethRPCURL string,
	chainID int64,
	entrypointAddress,
	beneficiaryAddress ethcmn.Address,
) (Entrypoint, error) {
	e := &entrypoint{
		entrypointAddress:  entrypointAddress,
		beneficiaryAddress: beneficiaryAddress,
	}

	ethClientRPC, err := ethrpc.Dial(ethRPCURL)
	if err != nil {
		err = errors.Wrap(err, "failed to dial eth RPC endpoint")
		return nil, err
	}

	e.rpcClient = ethClientRPC
	e.ethClient = ethclient.NewEthClient(ethClientRPC)
	e.chainID = big.NewInt(chainID)

	e.entrypointContract, err = contract_entrypoint.NewIEntryPoint(
		e.entrypointAddress,
		e.ethClient,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to wrap IEntryPoint contract")
		return nil, err
	}

	if e.entrypointABI, err = contract_entrypoint.IEntryPointMetaData.GetAbi(); err != nil {
		err = errors.Wrap(err, "failed to get IEntryPoint contract ABI bindings")
		return nil, err
	}

	rpcCallCtx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()

	// validate chainID against the RPC provided
	rpcChainID, err := e.ethClient.ChainID(rpcCallCtx)
	if err != nil {
		err = errors.Wrap(err, "failed to get chain ID from RPC")
		return nil, err
	}

	if rpcChainID.Cmp(e.chainID) != 0 {
		return nil, errors.Errorf("chain ID mismatch: expected %d, got %d", e.chainID, rpcChainID)
	}

	return e, nil
}

func (e *entrypoint) HandleOps(ctx context.Context, userOps []PackedUserOperation) (ethcmn.Hash, error) {
	txOpts := &bind.TransactOpts{
		Context: ctx,
	}

	packedUserOps := make([]contract_entrypoint.PackedUserOperation, len(userOps))
	for i, op := range userOps {
		packedUserOps[i] = contract_entrypoint.PackedUserOperation(op)
	}

	tx, err := e.entrypointContract.HandleOps(txOpts, packedUserOps, e.beneficiaryAddress)
	if err != nil {
		return ethcmn.Hash{}, errors.Wrap(err, "failed to handle user operations")
	}

	return tx.Hash(), nil
}

func (e *entrypoint) HandleOpsCallData(userOps []PackedUserOperation, beneficiaryAddress ethcmn.Address) ([]byte, error) {
	packedUserOps := make([]contract_entrypoint.PackedUserOperation, len(userOps))
	for i, op := range userOps {
		packedUserOps[i] = contract_entrypoint.PackedUserOperation(op)
	}

	callData, err := e.entrypointABI.Pack("handleOps", packedUserOps, beneficiaryAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack handleOps call data")
	}

	return callData, nil
}
