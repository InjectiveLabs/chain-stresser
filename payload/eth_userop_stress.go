package payload

import (
	"context"
	"math/big"
	"time"

	evmtypes "github.com/InjectiveLabs/sdk-go/chain/evm/types"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/xlab/pace"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chain-stresser/v2/eth/aa"
	contract "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/Counter"
)

var _ TxProvider = &ethUserOpProvider{}

type ethUserOpProvider struct {
	ethTxBuilderAndSigner

	minGasPrice          sdk.Coin
	maxGasLimitPerUserOp uint64
	maxGasLimitInitialTx uint64

	ethRPCURL string
	chainID   int64

	entrypoint            aa.Entrypoint
	entrypointAddress     ethcmn.Address
	accountFactoryAddress ethcmn.Address
	lightAccounts         map[int]aa.LightAccount
	lightAccountNonces    map[int]uint64

	counterContractMetaData *bind.MetaData
	counterContractABI      *abi.ABI
	counterContractAddress  ethcmn.Address

	logger       log.Logger
	uoSignedPace pace.Pace
}

const defaultAccountSalt = 1

// NewEthUserOpProvider creates transaction factory for stress testing
// Solidity contract transacting via UserOps (EIP-4337).
func NewEthUserOpProvider(
	ethRPCURL,
	chainID,
	minGasPrice string,
	uoSignedPace pace.Pace,
	entrypointAddress,
	beneficiaryAddress,
	accountFactoryAddress,
	counterContractAddress ethcmn.Address,
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

	provider := &ethUserOpProvider{
		ethTxBuilderAndSigner: ethTxBuilderAndSigner{
			ethSigner: ethSigner,
			feeDenom:  parsedMinGasPrice.Denom,
		},

		minGasPrice:          parsedMinGasPrice,
		maxGasLimitPerUserOp: 42000,
		maxGasLimitInitialTx: 230000,

		ethRPCURL:             ethRPCURL,
		chainID:               parsedChainID.Int64(),
		entrypointAddress:     entrypointAddress,
		accountFactoryAddress: accountFactoryAddress,

		lightAccounts:           make(map[int]aa.LightAccount),
		lightAccountNonces:      make(map[int]uint64),
		counterContractMetaData: contract.CounterMetaData,

		uoSignedPace: uoSignedPace,
	}

	entrypoint, err := aa.NewEntrypoint(
		ethRPCURL,
		parsedChainID.Int64(),
		accountFactoryAddress,
		beneficiaryAddress,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to create entrypoint")
		return nil, err
	} else {
		provider.entrypoint = entrypoint
	}

	contractABI, err := contract.CounterMetaData.GetAbi()
	if err != nil {
		err = errors.Wrap(err, "failed to parse Counter contract ABI")
		return nil, err
	} else {
		provider.counterContractABI = contractABI
	}

	provider.counterContractAddress = counterContractAddress

	provider.logger = log.WithFields(log.Fields{
		"provider": provider.Name(),
	})

	return provider, nil
}

type ethUserOpTx struct {
	baseTx

	to ethcmn.Address
}

func (p *ethUserOpProvider) Name() string {
	return "eth_userop_stress"
}

func (p *ethUserOpProvider) GenerateTx(
	req TxRequest,
) (Tx, error) {
	if p.counterContractAddress == (ethcmn.Address{}) {
		return nil, errors.New("counter contract address is not set")
	}

	counterCallData, err := p.counterContractABI.Pack("increase")
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack counter call calldata")
	}

	lightAccount, ok := p.lightAccounts[req.FromIdx]
	if !ok {
		return nil, errors.Errorf("light account not found for %d (%s)", req.FromIdx, req.From.Name)
	}

	personalSignFn := func(_ ethcmn.Address, data []byte) ([]byte, error) {
		pk, err := ethcrypto.ToECDSA([]byte(req.From.Key))
		if err != nil {
			return nil, err
		}

		digestHash := ethcrypto.Keccak256Hash(data)
		return ethcrypto.Sign(digestHash.Bytes(), pk)
	}

	signer := aa.NewSigner(
		p.entrypointAddress,
		p.ethSigner.ChainID(),
		personalSignFn,
		ethcmn.Address{}, // empty address ok since signerFn is fixed to privkey
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create AA signer for userOps")
	}

	// TODO: make this configurable
	opsPerTx := 50

	userOps := make([]aa.PackedUserOperation, opsPerTx)
	for i := 0; i < opsPerTx; i++ {
		lightAccountNonce := p.lightAccountNonces[req.FromIdx]

		userOp, err := lightAccount.NewUserOperationWithNonce(
			p.counterContractAddress,
			counterCallData,
			lightAccountNonce,
			aa.UserOperationGasEstimates{
				CallGasLimit:         22000,
				VerificationGasLimit: 75000,
				PreVerificationGas:   21000,
				MaxFeePerGas:         0,
				MaxPriorityFeePerGas: 0,
			},
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create user operation")
		}

		p.lightAccountNonces[req.FromIdx]++

		packedUO, _, err := signer.SignUserOperation(userOp)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to sign user operation (idx: %d, account: %d, nonce: %d)",
				i, req.FromIdx, lightAccountNonce,
			)
			return nil,
				err
		}

		userOps[i] = *packedUO
		p.uoSignedPace.StepN(1)
	}

	handleOpsCallData, err := p.entrypoint.HandleOpsCallData(userOps, req.From.EthAddress())
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack handleOps calldata")
	}

	tx := &ethUserOpTx{
		baseTx: baseTx{
			from: req.From,
			msgs: []sdk.Msg{
				evmtypes.NewTxWithData(&ethtypes.LegacyTx{
					Nonce:    req.From.Sequence,
					To:       &p.entrypointAddress,
					Value:    noValue,
					Gas:      p.maxGasLimitPerUserOp * uint64(len(userOps)),
					GasPrice: p.minGasPrice.Amount.BigInt(),
					Data:     handleOpsCallData,
				}),
			},

			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		},

		to: p.entrypointAddress,
	}

	return tx, nil
}

var (
	minAccountBalanceOnEndpoint, _ = big.NewInt(0).SetString("1000000000000000000", 10)
	endpointAccountBalanceTopup, _ = big.NewInt(0).SetString("1500000000000000000", 10)
)

func (p *ethUserOpProvider) GenerateInitialTx(
	req TxRequest,
) (Tx, error) {
	eoaAddress := req.From.EthAddress()

	la, err := aa.NewLightAccount(
		p.ethRPCURL,
		p.chainID,
		p.entrypointAddress,
		p.accountFactoryAddress,
		eoaAddress,
		defaultAccountSalt,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create light account (idx: %d; account: %s)", req.FromIdx, req.From.Name)
	}

	p.lightAccounts[req.FromIdx] = la

	initCtx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	nonce, err := la.Nonce(initCtx, ethcmn.Big0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync nonce for light account (idx: %d; account: %s)", req.FromIdx, req.From.Name)
	}

	p.lightAccountNonces[req.FromIdx] = nonce

	balance, err := la.Balance(initCtx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get balance for light account (idx: %d; account: %s)", req.FromIdx, req.From.Name)
	}

	if balance.Cmp(minAccountBalanceOnEndpoint) >= 0 {
		return nil, nil
	}

	depositCallData, err := la.DepositToEntrypointCallData()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get deposit to entrypoint call data for light account (idx: %d; account: %s)", req.FromIdx, req.From.Name)
	}

	tx := &baseTx{
		from: req.From,
		msgs: []sdk.Msg{
			evmtypes.NewTxWithData(&ethtypes.LegacyTx{
				Nonce:    req.From.Sequence,
				To:       &p.entrypointAddress,
				Value:    endpointAccountBalanceTopup,
				Gas:      p.maxGasLimitInitialTx,
				GasPrice: p.minGasPrice.Amount.BigInt(),
				Data:     depositCallData,
			}),
		},

		fromIdx: req.FromIdx,
		txIdx:   req.TxIdx,
	}

	return tx, nil
}
