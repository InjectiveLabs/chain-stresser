package payload

import (
	"context"
	_ "embed"
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

var (
	defaultMaxGasLimit uint64     = 10000000
	_                  TxProvider = &wasmInitContractProvider{}
)

type wasmInitContractProvider struct {
	queryClient chain.Client
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string

	contractCodeID uint64

	logger log.Logger
}

// NewWasmInitContractProvider creates transaction factory for stress testing
// wasm contract initialization.
func NewWasmInitContractProvider(
	queryClient chain.Client,
	minGasPrice string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &wasmInitContractProvider{
		queryClient: queryClient,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: defaultMaxGasLimit,
	}

	provider.logger = log.WithFields(log.Fields{
		"provider": provider.Name(),
	})

	return provider, nil
}

type wasmInitContractTx struct {
	baseTx
}

func (p *wasmInitContractProvider) Name() string {
	return "wasm_init_contract_stress"
}

const initMsgTemplate = `{
    "name": "CW20Solana",
    "symbol": "SOL",
    "decimals": 6,
    "initial_balances": [
        {
            "address": %q,
            "amount": "10000000000"
        }
    ],
    "mint": {
        "minter": %q
    },
    "marketing": {}
}`

func (p *wasmInitContractProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	sender := req.From.Key.AccAddress()

	msg := &wasmtypes.MsgInstantiateContract{
		Sender: sender,
		Admin:  sender,
		CodeID: p.contractCodeID,
		Label:  time.Now().Format(time.RFC3339Nano),
		Msg:    []byte(fmt.Sprintf(initMsgTemplate, sender, sender)),
		Funds: sdk.Coins{{
			Denom:  "inj",
			Amount: math.NewInt(1),
		}},
	}

	tx := &wasmInitContractTx{
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

func (p *wasmInitContractProvider) BuildAndSignTx(
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

func (p *wasmInitContractProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	if req.FromIdx != 0 || req.TxIdx != 0 {
		return nil, nil
	}

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

	if lastID, err := p.queryLastWasmContractID(p.queryClient); err != nil {
		return nil, errors.Wrap(err, "failed to query last CW contract ID")
	} else {
		p.contractCodeID = lastID + 1
	}

	p.logger.WithFields(log.Fields{
		"sender": sender,
		"codeID": p.contractCodeID,
	}).Infoln("Provisioned CW contract code ID")

	return tx, nil
}

func (p *wasmInitContractProvider) queryLastWasmContractID(client chain.Client) (uint64, error) {
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
