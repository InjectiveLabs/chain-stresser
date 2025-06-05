package payload

import (
	_ "embed"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var (
	//go:embed wasm/cw20_base.wasm
	cw20ByteCode []byte
	_            TxProvider = &wasmDeployProvider{}
)

type wasmDeployProvider struct {
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
}

// NewExchangeBatchUpdateProvider creates transaction factory for stress testing
// exchange batch updates.
func NewWasmDeployProvider(
	minGasPrice string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmDeployProvider{
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
	}

	return provider, nil
}

type wasmDeployTx struct {
	baseTx
}

func (p *wasmDeployProvider) Name() string {
	return "deploy_contract_stress"
}

func (p *wasmDeployProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	sender := req.From.Key.AccAddress()
	msg := &wasmtypes.MsgStoreCode{
		Sender:       sender,
		WASMByteCode: cw20ByteCode,
	}

	tx := &wasmDeployTx{
		baseTx: baseTx{
			from:    req.From,
			msgs:    []sdk.Msg{msg},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *wasmDeployProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	minGasPriceAmount := p.minGasPrice.Amount
	maxFeeAmount := minGasPriceAmount.Mul(math.NewIntFromUint64(p.maxGasLimit))

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

func (p *wasmDeployProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}
