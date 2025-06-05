package payload

import (
	"context"
	_ "embed"
	"fmt"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &wasmExecProvider{}

type wasmExecProvider struct {
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
}

// NewWasmExecProvider creates transaction factory for stress testing
// wasm contract execution.
func NewWasmExecProvider(
	minGasPrice string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmExecProvider{
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
	}

	return provider, nil
}

type execContractTx struct {
	baseTx
}

func (p *wasmExecProvider) Name() string {
	return "exec_contract_stress"
}

func (p *wasmExecProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	sender := req.From.Key.AccAddress()
	amount := math.NewInt(1000)
	msg := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: sender,
		Msg:      []byte(fmt.Sprintf(`{"mint":{"recipient": %q,"amount": "%q"}}`, sender, amount)),
		Funds: sdk.Coins{{
			Denom:  "inj",
			Amount: amount,
		}},
	}

	tx := &execContractTx{
		baseTx: baseTx{
			from:    req.From,
			msgs:    []sdk.Msg{msg},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *wasmExecProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	maxFeeAmount := p.minGasPrice.Amount.Mul(math.NewIntFromUint64(p.maxGasLimit))

	contractAddress, err := p.QueryLastWasmContractAddressBySender(client, unsignedTx.From().Key.AccAddress())
	if err != nil {
		return nil, err
	}

	msgs := unsignedTx.Msgs()
	msgIndex := unsignedTx.TxIdx()
	if msgIndex+1 >= len(msgs) {
		msgIndex = 0
	}

	msg := msgs[msgIndex]
	if typedMsg, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
		typedMsg.Contract = contractAddress
	}

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

func (p *wasmExecProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}

func (c *wasmExecProvider) QueryLastWasmContractAddressBySender(client chain.Client, address string) (string, error) {
	queryClient := client.NewWasmQueryClient()
	res, err := queryClient.ContractsByCreator(context.Background(), &wasmtypes.QueryContractsByCreatorRequest{CreatorAddress: address})
	if err != nil {
		return "", err
	}

	if l := len(res.ContractAddresses); l == 0 {
		return "", errors.New("no contract address found, please run tx-deploy-wasm-contract and tx-init-wasm-contract first")
	} else {
		return res.ContractAddresses[l-1], nil
	}
}
