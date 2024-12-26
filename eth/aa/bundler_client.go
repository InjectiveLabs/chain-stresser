package aa

import (
	"bytes"
	"context"
	"math/big"
	"net/http"
	"strings"
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	uint256type "github.com/holiman/uint256"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	ethclient "github.com/InjectiveLabs/chain-stresser/v2/eth/client"
	"github.com/InjectiveLabs/chain-stresser/v2/eth/jsonrpc"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

type BundlerClient interface {
	SendUserOp(ctx context.Context, userOp PackedUserOperation) error
}

func NewBundlerClient(
	ctx context.Context,
	aaBundlerUrl string,
	entrypointAddress ethcmn.Address,
	accountFactoryAddress ethcmn.Address,
) (BundlerClient, error) {
	bc := &bundlerClient{
		client: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},

		entrypointAddress:     entrypointAddress,
		accountFactoryAddress: accountFactoryAddress,
		bundlerUrl:            aaBundlerUrl,
		logger: log.DefaultLogger.WithFields(log.Fields{
			"svc": "aa_bundler",
		}),
	}

	ethClientRPC, err := ethrpc.Dial(aaBundlerUrl)
	if err != nil {
		err = errors.Wrap(err, "failed to dial bundler RPC endpoint")
		return nil, err
	}

	bc.ethClient = ethclient.NewEthClient(ethClientRPC)
	bc.rpcClient = ethClientRPC
	bc.rpcLoggingClient = jsonrpc.NewClientWithOpts(
		aaBundlerUrl,
		&jsonrpc.RPCClientOpts{
			AllowUnknownFields: true,
			LogEverything:      true,
			Logger:             bc.logger,
		},
	)

	rpcCallCtx, cancelFn := context.WithTimeout(ctx, 15*time.Second)
	defer cancelFn()

	// validate entrypoint address is supported by the RPC
	supportedEntryPoints, err := bc.supportedEntryPoints(rpcCallCtx)
	if err != nil {
		err = errors.Wrap(err, "failed to get supported entry points")
		return nil, err
	}

	isSupported := false
	for _, ep := range supportedEntryPoints {
		if ep == bc.entrypointAddress {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return nil, errors.Errorf("entrypoint address %s not supported by RPC", bc.entrypointAddress)
	}

	return bc, nil
}

type bundlerClient struct {
	client     *http.Client
	bundlerUrl string

	rpcClient        *ethrpc.Client
	rpcLoggingClient jsonrpc.RPCClient
	ethClient        ethclient.EthClientWithRet

	entrypointAddress     ethcmn.Address
	accountFactoryAddress ethcmn.Address

	logger log.Logger
}

func (c *bundlerClient) SendUserOp(ctx context.Context, userOp PackedUserOperation) error {
	return nil
}

func (s *bundlerClient) getTransactionReceipt(ctx context.Context, txHash ethcmn.Hash) (*ethtypes.Receipt, error) {
	var receipt *ethtypes.Receipt
	alchemyCallCtx, cancelFn := context.WithTimeout(ctx, 30*time.Second)
	err := s.rpcClient.CallContext(alchemyCallCtx, &receipt, "eth_getTransactionReceipt", txHash)
	cancelFn()

	if err != nil {
		err = errors.Wrap(err, "failed to call Bundler RPC to fetch Tx receipt")
		return nil, err
	}

	return receipt, nil
}

func (s *bundlerClient) getUserOperationReceipt(ctx context.Context, opHash ethcmn.Hash) (*bundlerUserOpReceipt, error) {
	var receipt *bundlerUserOpReceipt
	alchemyCallCtx, cancelFn := context.WithTimeout(ctx, 30*time.Second)
	err := s.rpcClient.CallContext(alchemyCallCtx, &receipt, "eth_getUserOperationReceipt", opHash)
	cancelFn()

	if err != nil {
		err = errors.Wrap(err, "failed to call Bundler RPC to fetch UserOperation receipt")
		return nil, err
	}

	return receipt, nil
}

func (s *bundlerClient) getUserOperationByHash(ctx context.Context, opHash ethcmn.Hash) (*bundlerUserOpResponse, error) {
	var response *bundlerUserOpResponse
	alchemyCallCtx, cancelFn := context.WithTimeout(ctx, 30*time.Second)
	err := s.rpcClient.CallContext(alchemyCallCtx, &response, "eth_getUserOperationByHash", opHash)
	cancelFn()

	if err != nil {
		err = errors.Wrap(err, "failed to call Bundler RPC to fetch UserOperation by hash")
		return nil, err
	}

	return response, nil
}

func (bc *bundlerClient) alchemyRequestGasAndPaymasterAndData(
	ctx context.Context,
	sender ethcmn.Address,
	nonce *big.Int,
	initCode,
	callData []byte,
) (estimates bundlerResponseGasAndPaymasterAndData, err error) {
	nonceU256 := uint256type.MustFromBig(nonce)

	ethCallCtx, cancelFn := context.WithTimeout(ctx, 20*time.Second)
	err = bc.rpcClient.CallContext(ethCallCtx, &estimates, "alchemy_requestGasAndPaymasterAndData",
		bundlerRequestGasAndPaymasterAndData{
			PolicyID:       "TODO_POLICY_ID",
			EntryPoint:     bc.entrypointAddress,
			DummySignature: gasEstimationDummySig,
			UserOperation: bundlerUserOperationEstimate{
				Sender:   sender,
				Nonce:    hexutil.U256(*nonceU256),
				InitCode: initCode,
				CallData: callData,
			},
			// Overrides: bundlerGasLimitOverrides{
			// 	MaxFeePerGas:         hexutil.U256(uint256type.NewInt(5000000)),
			// 	MaxPriorityFeePerGas: hexutil.U256(uint256type.NewInt(5000000)),
			// 	CallGasLimit:         hexutil.U256(uint256type.NewInt(5000000)),
			// 	VerificationGasLimit: hexutil.U256(uint256type.NewInt(5000000)),
			// 	PreVerificationGas:   hexutil.U256(uint256type.NewInt(5000000)),
			// },
		},
	)
	cancelFn()
	if err != nil {
		err = errors.Wrap(err, "failed to fetch UO gas esitmates from Bundler RPC")
		return bundlerResponseGasAndPaymasterAndData{}, err
	}

	return estimates, nil
}

func (bc *bundlerClient) sendUserOperation(
	ctx context.Context,
	userOperation *PackedUserOperation,
	userOperationHash ethcmn.Hash,
) (ethcmn.Hash, error) {
	var retUoHash hexutil.Bytes
	ethCallCtx, cancelFn := context.WithTimeout(ctx, 20*time.Second)
	err := bc.rpcClient.CallContext(ethCallCtx, &retUoHash, "eth_sendUserOperation", userOperation, bc.entrypointAddress)
	cancelFn()

	if err != nil {
		err = errors.Wrap(err, "failed to send UO via Bundler RPC")
		return userOperationHash, err
	} else if !bytes.Equal(retUoHash, userOperationHash.Bytes()) {
		bc.logger.WithFields(log.Fields{
			"uo_hash":         userOperationHash.String(),
			"alchemy_uo_hash": retUoHash.String(),
		}).Warningln("received different UO hash from Bundler RPC, replacing value")

		return ethcmn.BytesToHash(retUoHash), nil
	}

	return userOperationHash, nil
}

func isAA25NonceError(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), "AA25") {
		return true
	} else if strings.Contains(err.Error(), "invalid account nonce") {
		return true
	}

	return false
}

func (bc *bundlerClient) supportedEntryPoints(ctx context.Context) ([]ethcmn.Address, error) {
	var entryPoints []ethcmn.Address

	err := bc.rpcClient.CallContext(ctx, &entryPoints, "eth_supportedEntryPoints")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get supported entry points")
	}

	return entryPoints, nil
}

//
// DOCS: https://docs.stackup.sh/docs/useroperation-signature
//

const gasEstimationDummySig = "0xfffffffffffffffffffffffffffffff0000000000000000000000000000000007aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1c"

type bundlerUserOperationEstimate struct {
	Sender      ethcmn.Address `json:"sender"`
	Nonce       hexutil.U256   `json:"nonce"`
	Factory     hexutil.Bytes  `json:"factory,omitempty"`
	FactoryData hexutil.Bytes  `json:"factoryData,omitempty"`
	InitCode    hexutil.Bytes  `json:"initCode"`
	CallData    hexutil.Bytes  `json:"callData"`
}

type bundlerRequestGasAndPaymasterAndData struct {
	PolicyID       string                       `json:"policyId"`
	EntryPoint     ethcmn.Address               `json:"entryPoint"`
	DummySignature string                       `json:"dummySignature"`
	UserOperation  bundlerUserOperationEstimate `json:"userOperation"`
	Overrides      *bundlerGasLimitOverrides    `json:"overrides,omitempty"`
}

type bundlerGasLimitOverrides struct {
	MaxFeePerGas         hexutil.Bytes `json:"maxFeePerGas"`
	MaxPriorityFeePerGas hexutil.U256  `json:"maxPriorityFeePerGas"`
	CallGasLimit         hexutil.U256  `json:"callGasLimit"`
	VerificationGasLimit hexutil.U256  `json:"verificationGasLimit"`
	PreVerificationGas   hexutil.U256  `json:"preVerificationGas"`
}

type bundlerResponseGasAndPaymasterAndData struct {
	PaymasterAndData     hexutil.Bytes `json:"paymasterAndData"`
	CallGasLimit         hexutil.U256  `json:"callGasLimit"`
	VerificationGasLimit hexutil.U256  `json:"verificationGasLimit"`
	MaxPriorityFeePerGas hexutil.U256  `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         hexutil.U256  `json:"maxFeePerGas"`
	PreVerificationGas   hexutil.U256  `json:"preVerificationGas"`
}

type bundlerUserOpResponse struct {
	EntryPoint  ethcmn.Address `json:"entryPoint"`
	BlockNumber hexutil.Uint64 `json:"blockNumber"`
	BlockHash   ethcmn.Hash    `json:"blockHash"`
	TxHash      ethcmn.Hash    `json:"transactionHash"`
}

type bundlerUserOpReceipt struct {
	UserOpHash    ethcmn.Hash    `json:"userOpHash"`
	EntryPoint    ethcmn.Address `json:"entryPoint"`
	Sender        ethcmn.Address `json:"sender"`
	Nonce         hexutil.U256   `json:"nonce"`
	Paymaster     ethcmn.Address `json:"paymaster"`
	ActualGasCost hexutil.Big    `json:"actualGasCost"`
	ActualGasUsed hexutil.Uint64 `json:"actualGasUsed"`
	Success       bool           `json:"success"`
	Reason        string         `json:"reason"`

	TxReceipt ethtypes.Receipt `json:"receipt"`
}

func toBigInt(hexint hexutil.U256) *big.Int {
	return (*uint256type.Int)(&hexint).ToBig()
}

func ensureOnly0x(str string) string {
	if strings.HasPrefix(str, "0x0") {
		return "0x" + strings.TrimPrefix(str, "0x0")
	}

	return str
}
