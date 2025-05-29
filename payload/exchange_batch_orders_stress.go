package payload

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"cosmossdk.io/math"
	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

var _ TxProvider = &exchangeBatchOrdersProvider{}

type exchangeBatchOrdersProvider struct {
	spotMarketIDs       []string
	derivativeMarketIDs []string
	numTargets          int
	minGasPrice         sdk.Coin
	maxGasLimit         uint64
	memoAttach          string
}

// NewExchangeBatchUpdateProvider creates transaction factory for stress testing
// exchange batch updates.
func NewExchangeBatchOrdersProvider(
	minGasPrice string,
	spotMarketIDs []string,
	derivativeMarketIDs []string,
) (TxProvider, error) {

	parsedMinGasPrice, err := sdk.ParseCoinNormalized(minGasPrice)
	if err != nil {
		err = errors.Wrap(err, "failed to parse minGasPrice coin")
		return nil, err
	}

	provider := &exchangeBatchOrdersProvider{
		spotMarketIDs:       spotMarketIDs,
		derivativeMarketIDs: derivativeMarketIDs,
		minGasPrice:         parsedMinGasPrice,
		maxGasLimit:         150000,
	}

	return provider, nil
}

type exchangeBatchUpdateTx struct {
	baseTx
}

func (p *exchangeBatchOrdersProvider) Name() string {
	return "exchange_batch_orders_stress"
}

func (p *exchangeBatchOrdersProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	msg := new(exchangetypes.MsgBatchUpdateOrders)

	// Use standard random generation
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	price := math.LegacyNewDecFromInt(math.NewInt((r.Int63n(9) + 1) * 1000))
	quantity := math.LegacyNewDecFromIntWithPrec(math.NewInt(r.Int63n(10000)+1), 2)
	shouldCancelDerivative := r.Intn(2) > 0 // && false

	sender := req.From.Key.AccAddress()

	for i, marketID := range p.derivativeMarketIDs {
		price := price.Add(math.LegacyNewDec(int64(i * 1000)))

		derivativeOrder := &exchangetypes.DerivativeOrder{
			MarketId:  string(marketID),
			OrderType: exchangetypes.OrderType_BUY,
			Margin:    price.Mul(quantity),
			OrderInfo: exchangetypes.OrderInfo{
				FeeRecipient: sender,
				Price:        price,
				Quantity:     quantity,
				Cid:          time.Now().Format(time.RFC3339Nano),
			},
		}
		derivativeOrder.OrderInfo.SubaccountId = subaccount(req.From.Key.Address(), 0).Hex()

		msg.DerivativeOrdersToCreate = append(msg.DerivativeOrdersToCreate, derivativeOrder)
		if shouldCancelDerivative && len(p.derivativeMarketIDs) > 0 {
			// Cancel all orders for the current market ID
			msg.DerivativeMarketIdsToCancelAll = []string{string(marketID)}
			// Set SubaccountId to empty , TODO: check if need this at all
			msg.SubaccountId = subaccount(req.From.Key.Address(), 0).Hex()
		}
	}

	shouldCancelSpot := r.Intn(2) > 0 // && false
	for i, marketID := range p.spotMarketIDs {
		price := price.Add(math.LegacyNewDec(int64(i * 1000)))
		spotOrder := &exchangetypes.SpotOrder{
			MarketId:  string(marketID),
			OrderType: exchangetypes.OrderType_BUY,
			OrderInfo: exchangetypes.OrderInfo{
				FeeRecipient: sender,
				Price:        price,
				Quantity:     quantity,
				Cid:          time.Now().Format(time.RFC3339Nano),
			},
		}
		spotOrder.OrderInfo.SubaccountId = subaccount(req.From.Key.Address(), 0).Hex()
		msg.SpotOrdersToCreate = append(msg.SpotOrdersToCreate, spotOrder)
		if shouldCancelSpot {
			msg.SpotMarketIdsToCancelAll = []string{string(marketID)}
			msg.SubaccountId = subaccount(req.From.Key.Address(), 0).Hex()
		}
	}

	msg.Sender = sender

	tx := &exchangeBatchUpdateTx{
		baseTx: baseTx{
			from:    req.From,
			msgs:    []sdk.Msg{msg},
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *exchangeBatchOrdersProvider) BuildAndSignTx(
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

func (p *exchangeBatchOrdersProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}

func subaccount(account sdk.AccAddress, index int) eth.Hash {
	ethAddress := eth.BytesToAddress(account.Bytes())
	ethLowerAddress := strings.ToLower(ethAddress.String())

	subaccountId := fmt.Sprintf("%s%024x", ethLowerAddress, index)
	return eth.HexToHash(subaccountId)
}
