package payload

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"
	tokenfactorytypes "github.com/InjectiveLabs/sdk-go/chain/tokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &tokenFactoryBurnProvider{}

type tokenFactoryBurnProvider struct {
	burnAmount  math.Int
	mintAmount  math.Int
	minGasPrice sdk.Coin
	maxGasLimit uint64
	subdenom    string
}

// NewTokenFactoryBurnProvider creates a transaction factory that triggers
// tokenfactory MsgBurn, which burns from x/bank module balances internally.
func NewTokenFactoryBurnProvider(
	minGasPrice string,
	burnsPerAccount int,
) (TxProvider, error) {
	if burnsPerAccount <= 0 {
		return nil, errors.New("burnsPerAccount must be greater than 0")
	}

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &tokenFactoryBurnProvider{
		burnAmount:  math.NewInt(1),
		mintAmount:  math.NewInt(int64(burnsPerAccount) + 1),
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: 1_000_000,
		subdenom:    "burn" + strconv.FormatInt(time.Now().UnixNano(), 36),
	}

	return provider, nil
}

type tokenFactoryBurnTx struct {
	baseTx
}

func (p *tokenFactoryBurnProvider) Name() string {
	return "tokenfactory_burn_stress"
}

func (p *tokenFactoryBurnProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	fromAddr := req.From.Key.AccAddress()
	denom := p.denomFor(fromAddr)

	msg := &tokenfactorytypes.MsgBurn{
		Sender:          fromAddr,
		Amount:          sdk.NewCoin(denom, p.burnAmount),
		BurnFromAddress: fromAddr,
	}

	tx := &tokenFactoryBurnTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				msg,
			},

			provider: p,
			fromIdx:  req.FromIdx,
			txIdx:    req.TxIdx,
		},
	}

	return tx, nil
}

func (p *tokenFactoryBurnProvider) BuildAndSignTx(
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
	}

	signedResult, err := client.BuildAndSignTx(unsignedTx.From(), chainTx)
	if err != nil {
		return nil, err
	}

	tx := unsignedTx.WithBytes(client.Encode(signedResult))
	return tx, nil
}

func (p *tokenFactoryBurnProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	fromAddr := req.From.Key.AccAddress()
	denom := p.denomFor(fromAddr)

	tx := &tokenFactoryBurnTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				&tokenfactorytypes.MsgCreateDenom{
					Sender:         fromAddr,
					Subdenom:       p.subdenom,
					Name:           "Chain Stresser Burn",
					Symbol:         "CSB",
					Decimals:       0,
					AllowAdminBurn: true,
				},
				&tokenfactorytypes.MsgMint{
					Sender:   fromAddr,
					Amount:   sdk.NewCoin(denom, p.mintAmount),
					Receiver: fromAddr,
				},
			},

			provider: p,
			fromIdx:  req.FromIdx,
			txIdx:    req.TxIdx,
		},
	}

	return tx, nil
}

func (p *tokenFactoryBurnProvider) denomFor(creator string) string {
	return fmt.Sprintf("factory/%s/%s", creator, p.subdenom)
}
