package payload

import (
	"crypto/rand"
	"encoding/hex"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &bankMultiSendProvider{}

type bankMultiSendProvider struct {
	sendAmount  sdk.Coin
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
	numTargets  int
}

// NewBankMultiSendProvider creates transaction factory for stress testing
// native x/bank coin transfers to multiple accounts. Allows to inflate arity.
func NewBankMultiSendProvider(
	minGasPrice string,
	sendAmount string,
	numTargets int,
) (TxProvider, error) {
	if numTargets < 1 {
		return nil, errors.New("numTargets must be greater than 0")
	}

	parsedAmount, err := sdk.ParseCoinNormalized(sendAmount)
	if err != nil {
		err = errors.Wrap(err, "failed to parse amount coin")
		return nil, err
	}

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	randmemo := make([]byte, 128)
	_, err = rand.Read(randmemo)
	if err != nil {
		err = errors.Wrap(err, "failed to generate random memo")
		return nil, err
	}

	provider := &bankMultiSendProvider{
		sendAmount:  parsedAmount,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: 18000000, // 30000 * uint64(numTargets),
		memoAttach:  hex.EncodeToString(randmemo),
		numTargets:  numTargets,
	}

	return provider, nil
}

type bankMultiSendTx struct {
	baseTx

	to sdk.AccAddress
}

func (p *bankMultiSendProvider) Name() string {
	return "bank_multi_send_stress"
}

func (p *bankMultiSendProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	totalAmount := p.sendAmount.Amount.Mul(math.NewInt(int64(p.numTargets)))
	input := banktypes.NewInput(
		req.From.Key.Address(),
		sdk.Coins{
			sdk.NewCoin(
				p.sendAmount.Denom,
				totalAmount,
			),
		},
	)

	outputs := make([]banktypes.Output, 0, p.numTargets)

	for offset := 1; offset <= p.numTargets; offset++ {
		toIdx := req.FromIdx + offset
		if toIdx >= len(req.Keys) {
			toIdx = 0
		}

		outputs = append(outputs, banktypes.NewOutput(
			req.Keys[toIdx].Address(),
			sdk.Coins{
				p.sendAmount,
			},
		))
	}

	tx := &bankMultiSendTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				banktypes.NewMsgMultiSend(input, outputs),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},

		// doesn't matter for this tx
		to: sdk.AccAddress{},
	}

	return tx, nil
}

func (p *bankMultiSendProvider) BuildAndSignTx(
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

func (p *bankMultiSendProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}
