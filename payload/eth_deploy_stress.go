package payload

import (
	evmtypes "github.com/InjectiveLabs/sdk-go/chain/evm/types"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	contract "github.com/InjectiveLabs/chain-stresser/v2/contracts/solidity/Counter"
)

var _ TxProvider = &ethDeployProvider{}

type ethDeployProvider struct {
	ethTxBuilderAndSigner

	minGasPrice sdk.Coin
	maxGasLimit uint64

	contractMetaData *bind.MetaData
	contractABI      *abi.ABI
}

// NewEthDeployProvider creates transaction factory for stress testing
// Solidity contract transacting from multiple accounts.
func NewEthDeployProvider(
	chainID string,
	minGasPrice string,
) (TxProvider, error) {
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

	provider := &ethDeployProvider{
		ethTxBuilderAndSigner: ethTxBuilderAndSigner{
			ethSigner: ethSigner,
			feeDenom:  parsedMinGasPrice.Denom,
		},

		minGasPrice:      parsedMinGasPrice,
		maxGasLimit:      230000,
		contractMetaData: contract.CounterMetaData,
	}

	contractABI, err := contract.CounterMetaData.GetAbi()
	if err != nil {
		err = errors.Wrap(err, "failed to parse Counter contract ABI")
		return nil, err
	} else {
		provider.contractABI = contractABI
	}

	return provider, nil
}

type ethDeployTx struct {
	baseTx
}

func (p *ethDeployProvider) Name() string {
	return "eth_deploy_stress"
}

func (p *ethDeployProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	tx := &ethDeployTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				evmtypes.NewTxWithData(&ethtypes.LegacyTx{
					Nonce:    req.From.Sequence,
					To:       nil, // deployment
					Value:    noValue,
					Gas:      p.maxGasLimit,
					GasPrice: p.minGasPrice.Amount.BigInt(),
					Data:     ethcmn.FromHex(p.contractMetaData.Bin),
				}),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	return tx, nil
}

func (p *ethDeployProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	return nil, nil
}
