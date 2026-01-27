package payload

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &wasmExecContractProvider{}

type wasmExecContractProvider struct {
	queryClient chain.Client

	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string

	init2Salt        []byte
	contractCodeHash string
	contractCodeID   uint64
	contractAddress  string

	logger log.Logger
}

// NewWasmExecContractProvider creates transaction factory for stress testing
// wasm contract execution.
func NewWasmExecContractProvider(
	queryClient chain.Client,
	minGasPrice string,
) (TxProvider, error) {
	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmExecContractProvider{
		queryClient: queryClient,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
		init2Salt:   []byte(time.Now().Format(time.RFC3339Nano)),
	}

	codeHash, err := WasmCreateChecksum(cw20ByteCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wasm checksum for CW20 contract")
	}

	provider.contractCodeHash = codeHash.String()

	provider.logger = log.WithFields(log.Fields{
		"provider": provider.Name(),
	})

	return provider, nil
}

type wasmExecContractTx struct {
	baseTx
}

func (p *wasmExecContractProvider) Name() string {
	return "wasm_exec_contract_stress"
}

func (p *wasmExecContractProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	toIdx := req.FromIdx + 1
	if toIdx >= len(req.Keys) {
		toIdx = 0
	}

	to := req.Keys[toIdx].AccAddress()
	sender := req.From.Key.AccAddress()
	amount := 0 // mint is out of scope - not working

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: p.contractAddress,
		Msg:      []byte(fmt.Sprintf(`{"transfer":{"recipient": %q,"amount": "%d"}}`, to, amount)),
		Funds: sdk.Coins{{
			Denom:  "inj",
			Amount: math.NewInt(1),
		}},
	}

	tx := &wasmExecContractTx{
		baseTx: baseTx{
			from:     req.From,
			msgs:     []sdk.Msg{msg},
			provider: p,
			fromIdx:  req.FromIdx,
			txIdx:    req.TxIdx,
		},
	}

	return tx, nil
}

func (p *wasmExecContractProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	maxFeeAmount := p.minGasPrice.Amount.Mul(math.NewIntFromUint64(p.maxGasLimit))

	chainTx := chain.Tx{
		Msgs:     unsignedTx.Msgs(),
		GasLimit: p.maxGasLimit,
		Fee: sdk.NewCoins(sdk.NewCoin(
			p.minGasPrice.Denom,
			maxFeeAmount,
		)),
		Memo: p.memoAttach,
	}

	signedResult, err := client.BuildAndSignTx(unsignedTx.From(), chainTx)
	if err != nil {
		return nil, err
	}

	tx := unsignedTx.WithBytes(client.Encode(signedResult))
	return tx, nil
}

func (p *wasmExecContractProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	if req.FromIdx != 0 || req.TxIdx != 0 {
		return nil, nil
	}

	sender := req.From.Key.AccAddress()
	initMsgBytes := []byte(fmt.Sprintf(initMsgTemplate, sender, sender))

	if lastID, err := p.queryLastWasmContractID(p.queryClient); err != nil {
		return nil, errors.Wrap(err, "failed to query last CW contract ID")
	} else {
		p.contractCodeID = lastID + 1
	}

	if contractAddress, err := p.queryBuildAddress(
		p.queryClient,
		req.From.Key.AccAddress(),
		p.contractCodeHash,
		initMsgBytes,
	); err != nil {
		return nil, errors.Wrap(err, "failed to query build address")
	} else {
		p.contractAddress = contractAddress
	}

	msgStore := &wasmtypes.MsgStoreCode{
		Sender:       sender,
		WASMByteCode: cw20ByteCode,
	}

	msgInstantiate2 := &wasmtypes.MsgInstantiateContract2{
		Sender: sender,
		CodeID: p.contractCodeID,
		Label:  time.Now().Format(time.RFC3339Nano),
		Msg:    initMsgBytes,
		Funds: sdk.Coins{{
			Denom:  "inj",
			Amount: math.NewInt(1),
		}},
		Salt:   p.init2Salt,
		FixMsg: true,
	}

	tx := &wasmStoreCodeTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				msgStore,
				msgInstantiate2,
			},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	p.logger.WithFields(log.Fields{
		"sender":  sender,
		"codeID":  p.contractCodeID,
		"address": p.contractAddress,
	}).Infoln("Provisioned CW contract")

	return tx, nil
}

func (p *wasmExecContractProvider) queryLastWasmContractID(client chain.Client) (uint64, error) {
	queryClient := client.NewWasmQueryClient()

	totalCodes := 0
	nextKey := []byte{}
	for {
		res, err := queryClient.Codes(context.Background(),
			&wasmtypes.QueryCodesRequest{
				Pagination: &query.PageRequest{
					Limit: 100, // This is highest value allowed by the API
					Key:   nextKey,
				},
			},
		)
		if err != nil {
			return 0, errors.Wrap(err, "failed to query CW codes")
		}

		totalCodes += len(res.CodeInfos)
		if len(res.Pagination.NextKey) > 0 {
			nextKey = res.Pagination.NextKey
			continue
		}

		return uint64(totalCodes), nil
	}
}

func (p *wasmExecContractProvider) queryBuildAddress(
	client chain.Client,
	sender string,
	codeHash string,
	initArgs []byte,
) (string, error) {
	queryClient := client.NewWasmQueryClient()

	res, err := queryClient.BuildAddress(context.Background(),
		&wasmtypes.QueryBuildAddressRequest{
			CodeHash:       codeHash,
			CreatorAddress: sender,
			Salt:           hex.EncodeToString(p.init2Salt),
			InitArgs:       initArgs,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to query build address")
	}

	return res.Address, nil
}
