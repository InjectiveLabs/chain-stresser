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

	contract "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/BenchmarkInternalCall"
)

var _ TxProvider = &ethInternalCallProvider{}

type ethInternalCallProvider struct {
	ethTxBuilderAndSigner

	minGasPrice           sdk.Coin
	maxGasLimit           uint64
	maxGasLimitDeployment uint64
	iterations            uint64

	contractMetaData *bind.MetaData
	contractABI      *abi.ABI
	contractAddress  ethcmn.Address

	logger log.Logger
}

const defaultEthInternalCallIterations = 100

// NewEthInternalCallProvider creates transaction factory for stress testing
// Solidity contract calling internal contract.
func NewEthInternalCallProvider(
	chainID string,
	minGasPrice string,
	iterations uint64,
) (TxProvider, error) {
	if iterations == 0 {
		iterations = defaultEthInternalCallIterations
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

	provider := &ethInternalCallProvider{
		ethTxBuilderAndSigner: ethTxBuilderAndSigner{
			ethSigner: ethSigner,
			feeDenom:  parsedMinGasPrice.Denom,
		},

		minGasPrice:           parsedMinGasPrice,
		maxGasLimit:           28809 + 2500*iterations,
		maxGasLimitDeployment: 320000,
		contractMetaData:      contract.BenchmarkInternalCallMetaData,
		iterations:            iterations,
	}

	contractABI, err := contract.BenchmarkInternalCallMetaData.GetAbi()
	if err != nil {
		err = errors.Wrap(err, "failed to parse BenchmarkInternalCall contract ABI")
		return nil, err
	} else {
		provider.contractABI = contractABI
	}

	provider.logger = log.WithFields(log.Fields{
		"provider": provider.Name(),
	})

	return provider, nil
}

type ethInternalCallTx struct {
	baseTx

	to ethcmn.Address
}

func (p *ethInternalCallProvider) Name() string {
	return "eth_internal_call_stress"
}

func (p *ethInternalCallProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	if p.contractAddress == (ethcmn.Address{}) {
		return nil, errors.New("contract address is not set")
	}

	// run the heavy call!
	callData, err := p.contractABI.Pack("benchmarkInternalCall", big.NewInt(int64(p.iterations)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack contract call data")
	}

	tx := &ethInternalCallTx{
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

func (p *ethInternalCallProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	if req.FromIdx != 0 || req.TxIdx != 0 {
		return nil, nil
	}

	tx := &baseTx{
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
	}

	ethFrom := ethcmn.BytesToAddress(req.From.Key.PubKey().Address().Bytes())
	p.contractAddress = ethcrypto.CreateAddress(ethFrom, req.From.Sequence)
	p.logger.WithField("address", p.contractAddress.String()).Infoln("Provisioned BenchmarkInternalCall contract address")

	return tx, nil
}
