package payload

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var (
	defaultMaxGasLimit uint64     = 5000000
	_                  TxProvider = &wasmInitProvider{}
)

type wasmInitProvider struct {
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
}

// NewWasmInitProvider creates transaction factory for stress testing
// wasm contract initialization.
func NewWasmInitProvider(
	minGasPrice string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmInitProvider{
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
	}

	return provider, nil
}

type wasmInitTx struct {
	baseTx
}

func (p *wasmInitProvider) Name() string {
	return "init_contract_stress"
}

func (p *wasmInitProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	sender := req.From.Key.AccAddress()

	msg := &wasmtypes.MsgInstantiateContract{
		Sender: sender,
		Admin:  sender,
		CodeID: 99, // Set arbitarry codeID that will be replaced in BuildAndSignTx
		Label:  time.Now().Format(time.RFC3339Nano),
		Msg:    []byte(fmt.Sprintf(`{"name":"CW20Solana","symbol":"SOL","decimals":6,"initial_balances":[{"address":%q,"amount":"10000000000"}],"mint":{"minter":%q},"marketing":{}}`, sender, sender)),
		Funds: sdk.Coins{{
			Denom:  "inj",
			Amount: math.NewInt(1),
		}},
	}

	fmt.Println("msg.CodeID in GenerateTx", msg.CodeID)

	tx := &wasmInitTx{
		baseTx: baseTx{
			from:    req.From,
			msgs:    []sdk.Msg{msg},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *wasmInitProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	minGasPriceAmount := p.minGasPrice.Amount
	maxFeeAmount := minGasPriceAmount.Mul(math.NewIntFromUint64(p.maxGasLimit))

	msgs := unsignedTx.Msgs()
	msgIndex := unsignedTx.TxIdx()
	if msgIndex+1 >= len(msgs) {
		msgIndex = 0
	}

	msg := msgs[msgIndex]
	if typedMsg, ok := msg.(*wasmtypes.MsgInstantiateContract); ok {
		// Get code id
		codeID, err := p.QueryLastWasmCodeIDBySender(client, typedMsg.Sender)
		if err != nil {
			return nil, err
		}
		if codeID != 0 {
			typedMsg.CodeID = codeID
		}

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

func (p *wasmInitProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}

func (p *wasmInitProvider) QueryLastWasmCodeIDBySender(client chain.Client, sender string) (uint64, error) {
	queryClient := client.NewWasmQueryClient()

	res, err := queryClient.Codes(context.Background(), &wasmtypes.QueryCodesRequest{})
	if err != nil {
		return 0, err
	}
	for i := len(res.CodeInfos) - 1; i >= 0; i-- {
		code := res.CodeInfos[i]
		if code.Creator == sender {
			return code.CodeID, nil
		}
	}
	return 0, nil
}
