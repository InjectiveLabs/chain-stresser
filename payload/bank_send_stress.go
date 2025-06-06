package payload

import (
	"encoding/hex"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &bankSendProvider{}

type bankSendProvider struct {
	sendAmount  sdk.Coin
	minGasPrice sdk.Coin
	maxGasLimit uint64
	memoAttach  string
}

// NewBankSendProvider creates transaction factory for stress testing
// native x/bank coin transfers between accounts.
func NewBankSendProvider(
	minGasPrice string,
	sendAmount string,
) (TxProvider, error) {
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

	var randmemo []byte
	// randmemo := make([]byte, 128)
	// _, err = rand.Read(randmemo)
	// if err != nil {
	// 	err = errors.Wrap(err, "failed to generate random memo")
	// 	return nil, err
	// }

	provider := &bankSendProvider{
		sendAmount:  parsedAmount,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: 150000,
		memoAttach:  hex.EncodeToString(randmemo),
	}

	return provider, nil
}

type bankSendTx struct {
	baseTx

	to sdk.AccAddress
}

func (p *bankSendProvider) Name() string {
	return "bank_send_stress"
}

func (p *bankSendProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	toIdx := req.FromIdx + 1
	if toIdx >= len(req.Keys) {
		toIdx = 0
	}

	sendCoins := sdk.Coins{
		p.sendAmount,
	}

	to := req.Keys[toIdx].Address()
	tx := &bankSendTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(req.From.Key.Address(), to, sendCoins),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},

		to: to,
	}

	return tx, nil
}

func (p *bankSendProvider) BuildAndSignTx(
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

func (p *bankSendProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}
