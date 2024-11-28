package payload

import (
	"math/big"

	evmtypes "github.com/InjectiveLabs/sdk-go/chain/evm/types"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	contract "github.com/InjectiveLabs/chain-stresser/v2/contracts/solidity/Counter"
)

var _ TxProvider = &ethCallProvider{}

type ethCallProvider struct {
	ethTxBuilderAndSigner

	logger log.Logger

	minGasPrice           sdk.Coin
	maxGasLimit           uint64
	maxGasLimitDeployment uint64

	contractMetaData *bind.MetaData
	contractABI      *abi.ABI
	contractAddress  ethcmn.Address
}

// NewEthCallProvider creates transaction factory for stress testing
// Solidity contract transacting from multiple accounts.
func NewEthCallProvider(
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

	provider := &ethCallProvider{
		ethTxBuilderAndSigner: ethTxBuilderAndSigner{
			ethSigner: ethSigner,
			feeDenom:  parsedMinGasPrice.Denom,
		},

		minGasPrice:           parsedMinGasPrice,
		maxGasLimit:           50000,
		maxGasLimitDeployment: 230000,
		contractMetaData:      contract.CounterMetaData,
	}

	contractABI, err := contract.CounterMetaData.GetAbi()
	if err != nil {
		err = errors.Wrap(err, "failed to parse Counter contract ABI")
		return nil, err
	} else {
		provider.contractABI = contractABI
	}

	provider.logger = log.WithFields(log.Fields{
		"provider": provider.Name(),
	})

	return provider, nil
}

type ethCallTx struct {
	baseTx

	to ethcmn.Address
}

func (p *ethCallProvider) Name() string {
	return "eth_call_stress"
}

var noValue = big.NewInt(0)

func (p *ethCallProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	if p.contractAddress == (ethcmn.Address{}) {
		return nil, errors.New("contract address is not set")
	}

	// increase the counter!
	callData, err := p.contractABI.Pack("increase")
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack contract call data")
	}

	tx := &ethCallTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				evmtypes.NewTxWithData(&ethtypes.LegacyTx{
					Nonce:    req.From.Sequence,
					To:       &p.contractAddress,
					Value:    noValue,
					Gas:      p.maxGasLimit,
					GasPrice: p.minGasPrice.Amount.BigInt(),
					Data:     callData,
				}),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},

		to: p.contractAddress,
	}

	return tx, nil
}

func (p *ethCallProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	tx := &ethCallTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				evmtypes.NewTxWithData(&ethtypes.LegacyTx{
					Nonce:    req.From.Sequence,
					To:       nil, // contract deployment
					Value:    noValue,
					Gas:      p.maxGasLimitDeployment,
					GasPrice: p.minGasPrice.Amount.BigInt(),
					Data:     ethcmn.FromHex(p.contractMetaData.Bin),
				}),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},
	}

	ethFrom := ethcmn.BytesToAddress(req.From.Key.PubKey().Address().Bytes())
	p.contractAddress = ethcrypto.CreateAddress(ethFrom, req.From.Sequence)
	p.logger.WithField("address", p.contractAddress.String()).Infoln("Provisioned Counter contract address")

	return tx, nil
}
