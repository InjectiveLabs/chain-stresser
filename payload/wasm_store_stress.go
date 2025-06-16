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
	_            TxProvider = &wasmStoreCodeProvider{}
)

type wasmStoreCodeProvider struct {
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
}

// NewWasmStoreCodeProvider creates transaction factory for stress testing
// wasm contract code storement.
func NewWasmStoreCodeProvider(
	minGasPrice string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmStoreCodeProvider{
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
	}

	return provider, nil
}

type wasmStoreCodeTx struct {
	baseTx
}

func (p *wasmStoreCodeProvider) Name() string {
	return "wasm_store_code_stress"
}

func (p *wasmStoreCodeProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	sender := req.From.Key.AccAddress()
	msg := &wasmtypes.MsgStoreCode{
		Sender:       sender,
		WASMByteCode: cw20ByteCode,
	}

	tx := &wasmStoreCodeTx{
		baseTx: baseTx{
			from:    req.From,
			msgs:    []sdk.Msg{msg},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *wasmStoreCodeProvider) BuildAndSignTx(
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

func (p *wasmStoreCodeProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	// Not implemented, because wasm code store doesn't require any prior state on the chain

	return nil, nil
}
