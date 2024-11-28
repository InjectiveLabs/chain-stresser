package payload

import (
	"cosmossdk.io/math"
	evmtypes "github.com/InjectiveLabs/sdk-go/chain/evm/types"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

var _ TxProvider = &ethSendProvider{}

type ethSendProvider struct {
	ethTxBuilderAndSigner

	sendAmount  sdk.Coin
	minGasPrice sdk.Coin
	maxGasLimit uint64
}

// eip1559InitiaBaseFee defines the initial base fee for EIP-1559 transactions.
var eip1559InitialBaseFee = math.NewIntFromUint64(1000000000)

// NewEthSendProvider creates transaction factory for stress testing
// native eth transfers (EVM -> x/bank) between accounts.
func NewEthSendProvider(
	chainID string,
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

	// override minGasPrice if it's less than the initial base fee for EIP-1559 transactions
	if parsedMinGasPrice.Amount.LT(eip1559InitialBaseFee) {
		parsedMinGasPrice.Amount = eip1559InitialBaseFee
	}

	parsedChainID, err := chaintypes.ParseChainID(chainID)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse chainID: %s", chainID)
		return nil, err
	}

	ethSigner := ethtypes.LatestSignerForChainID(parsedChainID)

	provider := &ethSendProvider{
		ethTxBuilderAndSigner: ethTxBuilderAndSigner{
			ethSigner: ethSigner,
			feeDenom:  parsedMinGasPrice.Denom,
		},

		sendAmount:  parsedAmount,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: 21000,
	}

	return provider, nil
}

type ethSendTx struct {
	baseTx

	to sdk.AccAddress
}

func (p *ethSendProvider) Name() string {
	return "eth_send_stress"
}

func (p *ethSendProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	toIdx := req.FromIdx + 1
	if toIdx >= len(req.Keys) {
		toIdx = 0
	}

	to := req.Keys[toIdx].Address()
	toEth := ethcmn.Address(to.Bytes())

	tx := &ethSendTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				evmtypes.NewTxWithData(&ethtypes.LegacyTx{
					Nonce:    req.From.Sequence,
					To:       &toEth,
					Value:    p.sendAmount.Amount.BigInt(),
					Gas:      p.maxGasLimit,
					GasPrice: p.minGasPrice.Amount.BigInt(),
					Data:     nil, // simple value transfer
				}),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},

		to: to,
	}

	return tx, nil
}

func (p *ethSendProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}
