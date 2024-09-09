package payload

import (
	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
)

var _ TxProvider = &ethSendProvider{}

type ethSendProvider struct {
	sendAmount  sdk.Coin
	minGasPrice sdk.Coin
	maxGasLimit uint64
	ethSigner   ethtypes.Signer
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

	parsedChainID, err := ethermint.ParseChainID(chainID)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse chainID: %s", chainID)
		return nil, err
	}

	ethSigner := ethtypes.LatestSignerForChainID(parsedChainID)

	provider := &ethSendProvider{
		sendAmount:  parsedAmount,
		minGasPrice: parsedMinGasPrice,
		maxGasLimit: 200000,
		ethSigner:   ethSigner,
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

func (p *ethSendProvider) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	ethTxMsg := unsignedTx.Msgs()[0].(*evmtypes.MsgEthereumTx)
	txHash := p.ethSigner.Hash(ethTxMsg.Raw.Transaction)

	privKey := &ethsecp256k1.PrivKey{
		Key: unsignedTx.From().Key,
	}

	sig, err := ethcrypto.Sign(txHash.Bytes(), privKey.ToECDSA())
	if err != nil {
		err = errors.Wrap(err, "failed to sign tx hash")
		return nil, err
	}

	ethTxMsg.Raw.Transaction, err =
		ethTxMsg.Raw.Transaction.WithSignature(p.ethSigner, sig)
	if err != nil {
		err = errors.Wrap(err, "failed to update tx with signature")
		return nil, err
	}

	ethTxMsg.From = privKey.PubKey().Address().Bytes()

	txBuilder := client.TxConfig().NewTxBuilder()

	ethTxOpt, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		err := errors.New("failed to init NewAnyWithValue for ExtensionOptionsEthereumTx")
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		err := errors.New("txBuilder isn't authtx.ExtensionOptionsTxBuilder")
		return nil, err
	} else {
		builder.SetExtensionOptions(ethTxOpt)
	}

	tx, err := ethTxMsg.BuildTx(builder, p.minGasPrice.Denom)
	if err != nil {
		err = errors.Wrap(err, "failed to build MsgEthereumTx")
		return nil, err
	}

	out := *(unsignedTx.(*ethSendTx))
	out.txBytes = client.Encode(tx)
	return &out, nil
}
