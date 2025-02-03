package aa

import (
	"context"
	"encoding/hex"
	"math/big"
	"time"

	ethclient "github.com/InjectiveLabs/chain-stresser/v2/eth/client"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	contract_entrypoint "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/IEntrypoint"
	contract_light_account "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/LightAccount"
	contract_light_account_factory "github.com/InjectiveLabs/chain-stresser/v2/eth/solidity/LightAccountFactory"
)

type LightAccount interface {
	Nonce(ctx context.Context, key *big.Int) (uint64, error)
	Address(ctx context.Context) (ethcmn.Address, error)
	Balance(ctx context.Context) (*big.Int, error)
	Code(ctx context.Context) ([]byte, error)

	NewUserOperation(
		ctx context.Context,
		contractAddress ethcmn.Address,
		callData []byte,
		gasEstimates ...UserOperationGasEstimates,
	) (*UserOperation, error)

	NewUserOperationWithNonce(
		contractAddress ethcmn.Address,
		callData []byte,
		nonce uint64,
		gasEstimates ...UserOperationGasEstimates,
	) (*UserOperation, error)

	DepositToEntrypointCallData() ([]byte, error)
}

func NewLightAccount(
	ethRPCURL string,
	chainID int64,
	entrypointAddress ethcmn.Address,
	accountFactoryAddress ethcmn.Address,
	eoaOwner ethcmn.Address,
	salt uint64,
) (LightAccount, error) {
	la := &lightAccount{
		entrypointAddress:     entrypointAddress,
		accountFactoryAddress: accountFactoryAddress,
		eoaOwner:              eoaOwner,
		salt:                  salt,
	}

	ethClientRPC, err := ethrpc.Dial(ethRPCURL)
	if err != nil {
		err = errors.Wrap(err, "failed to dial eth RPC endpoint")
		return nil, err
	}

	la.ethClient = ethclient.NewEthClient(ethClientRPC)

	la.accountFactoryContract, err = contract_light_account_factory.NewLightAccountFactory(
		accountFactoryAddress,
		la.ethClient,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to wrap LightAccountFactory contract")
		return nil, err
	}

	if la.accountFactoryABI, err = contract_light_account_factory.LightAccountFactoryMetaData.GetAbi(); err != nil {
		err = errors.Wrap(err, "failed to get LightAccountFactory contract ABI bindings")
		return nil, err
	}

	la.accountContract, err = contract_light_account.NewLightAccount(
		la.accountAddress,
		la.ethClient,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to wrap LightAccount contract")
		return nil, err
	}

	if la.accountABI, err = contract_light_account.LightAccountMetaData.GetAbi(); err != nil {
		err = errors.Wrap(err, "failed to get LightAccount contract ABI bindings")
		return nil, err
	}

	la.entrypointContract, err = contract_entrypoint.NewIEntryPoint(
		entrypointAddress,
		la.ethClient,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to wrap IEntryPoint contract")
		return nil, err
	}

	if la.entrypointABI, err = contract_entrypoint.IEntryPointMetaData.GetAbi(); err != nil {
		err = errors.Wrap(err, "failed to get IEntryPoint contract ABI bindings")
		return nil, err
	}

	initCtx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()

	la.chainID = big.NewInt(chainID)

	// validate chainID against the RPC provided
	rpcChainID, err := la.ethClient.ChainID(initCtx)
	if err != nil {
		err = errors.Wrap(err, "failed to get chain ID from RPC")
		return nil, err
	}

	if rpcChainID.Cmp(la.chainID) != 0 {
		return nil, errors.Errorf("chain ID mismatch: expected %d, got %d", la.chainID, rpcChainID)
	}

	la.logger = log.WithFields(log.Fields{
		"svc":     "light_account",
		"factory": la.accountFactoryAddress.Hex(),
		"eoa":     la.eoaOwner.Hex(),
		"chainID": chainID,
	})

	// get the address of the light account
	if _, err = la.Address(initCtx); err != nil {
		err = errors.Wrap(err, "post-check after light account init: failed to get address (even optimistically)")
		return nil, err
	}

	la.logger = la.logger.WithFields(log.Fields{
		"address": la.accountAddress.Hex(),
	})

	if salt != 0 {
		la.logger = la.logger.WithFields(log.Fields{
			"salt": la.salt,
		})
	}

	// check if the account is already deployed
	if isDeployed, err := la.IsDeployed(initCtx); err != nil {
		err = errors.Wrap(err, "post-check after light account init: failed to check if account is deployed")
		return nil, err
	} else if isDeployed {
		la.logger.Debugln("light account is already deployed")
	}

	return la, nil
}

type lightAccount struct {
	ethClient ethclient.EthClientWithRet
	chainID   *big.Int

	entrypointAddress     ethcmn.Address
	accountFactoryAddress ethcmn.Address
	accountAddress        ethcmn.Address
	eoaOwner              ethcmn.Address
	salt                  uint64
	isDeployed            bool

	accountFactoryABI      *abi.ABI
	accountFactoryContract *contract_light_account_factory.LightAccountFactory
	accountABI             *abi.ABI
	accountContract        *contract_light_account.LightAccount

	entrypointABI      *abi.ABI
	entrypointContract *contract_entrypoint.IEntryPoint

	logger log.Logger
}

func (la *lightAccount) Nonce(ctx context.Context, key *big.Int) (uint64, error) {
	res, err := la.entrypointContract.GetNonce(&bind.CallOpts{
		Context: ctx,
	}, la.accountAddress, key)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get nonce")
	}

	return res.Uint64(), nil
}

func (la *lightAccount) Address(ctx context.Context) (ethcmn.Address, error) {
	if (la.accountAddress != ethcmn.Address{}) {
		return la.accountAddress, nil
	}

	res, err := la.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		From: la.eoaOwner,
		To:   &la.accountFactoryAddress,
		Data: la.createAccountCallData(la.eoaOwner, la.salt),
	}, nil)

	if err != nil {
		return ethcmn.Address{}, errors.Wrap(err, "failed to read light account address from factory")
	}

	la.accountAddress = ethcmn.BytesToAddress(res)

	la.logger.WithFields(log.Fields{
		"res":  hex.EncodeToString(res),
		"eoa":  la.eoaOwner.Hex(),
		"salt": la.salt,
	}).Debugf("received light account address: %s", la.accountAddress.Hex())

	return la.accountAddress, nil
}

func (la *lightAccount) Balance(ctx context.Context) (*big.Int, error) {
	balance, err := la.entrypointContract.BalanceOf(&bind.CallOpts{
		Context: ctx,
	}, la.eoaOwner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance")
	}

	return balance, nil
}

func (la *lightAccount) DepositToEntrypointCallData() ([]byte, error) {
	callData, err := la.entrypointABI.Pack("depositTo", la.eoaOwner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack depositToEntrypoint call data")
	}

	return callData, nil
}

func (la *lightAccount) Code(ctx context.Context) ([]byte, error) {
	if (la.accountAddress == ethcmn.Address{}) {
		return nil, errors.New("light account address is not set (it has to be derived from factory)")
	}

	codeBytes, err := la.ethClient.CodeAt(ctx, la.accountAddress, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get code")
	}

	return codeBytes, nil
}

func (la *lightAccount) IsDeployed(ctx context.Context) (bool, error) {
	if la.isDeployed {
		// already checked, cannot undeploy
		return true, nil
	}

	codeBytes, err := la.Code(ctx)
	if err != nil {
		return false, errors.WithStack(err)
	}

	la.isDeployed = len(codeBytes) > 0

	return la.isDeployed, nil
}

func (la *lightAccount) NewUserOperation(
	ctx context.Context,
	contractAddress ethcmn.Address,
	callData []byte,
	gasEstimates ...UserOperationGasEstimates,
) (*UserOperation, error) {
	dynamicNonce, err := la.Nonce(ctx, ethcmn.Big0)
	if err != nil {
		la.logger.WithError(err).Errorln("failed to get account nonce")
		return nil, errors.Wrap(err, "failed to get account nonce")
	}

	return la.newUserOperation(
		contractAddress,
		callData,
		dynamicNonce,
		gasEstimates...,
	)
}

func (la *lightAccount) NewUserOperationWithNonce(
	contractAddress ethcmn.Address,
	callData []byte,
	nonce uint64,
	gasEstimates ...UserOperationGasEstimates,
) (*UserOperation, error) {
	return la.newUserOperation(
		contractAddress,
		callData,
		nonce,
		gasEstimates...,
	)
}

func (la *lightAccount) initCode() []byte {
	factoryCallData := la.createAccountCallData(la.eoaOwner, la.salt)
	initCode := append([]byte{}, la.accountFactoryAddress.Bytes()...)
	initCode = append(initCode, factoryCallData...)
	return initCode
}

func (la *lightAccount) createAccountCallData(eoaOwner ethcmn.Address, salt uint64) []byte {
	saltBig := ethcmn.Big0
	if salt != 0 {
		saltBig = big.NewInt(int64(la.salt))
	}

	callData, err := la.accountFactoryABI.Pack(
		"createAccount",
		eoaOwner,
		saltBig,
	)
	if err != nil {
		panic(errors.Wrap(err, "failed to pack factory data"))
	}

	return callData
}

func (la *lightAccount) getNonceCallData() []byte {
	callData, err := la.accountABI.Pack("getNonce")
	if err != nil {
		panic(errors.Wrap(err, "failed to pack calldata for LightAccount getNonce method"))
	}

	return callData
}

func (la *lightAccount) executeCallData(contractAddress ethcmn.Address, callData []byte) []byte {
	result, err := la.accountABI.Pack("execute", contractAddress, ethcmn.Big0, callData)
	if err != nil {
		panic(errors.Wrap(err, "failed to pack calldata for LightAccount execute method"))
	}

	return result
}

// newUserOperation constructs a new UO to be sent for the light account.
// ultimate target contract is specified by contractAddress
// ultimate useful payload is specified by callData
// method allows to specify nonce and gas estimates
func (la *lightAccount) newUserOperation(
	contractAddress ethcmn.Address,
	callData []byte,
	nonce uint64,
	gasEstimates ...UserOperationGasEstimates,
) (*UserOperation, error) {
	uoLogger := la.logger.WithFields(log.Fields{
		"isDeployed":  la.isDeployed,
		"contract":    contractAddress.Hex(),
		"callDataLen": len(callData),
		"nonce":       nonce,
	})

	if len(gasEstimates) == 0 {
		uoLogger.Warningln("no gas estimates provided, cannot estimate gas using AA API yet")
		return nil, errors.New("no gas estimates provided")
	}

	executeCallData := la.executeCallData(
		contractAddress, // ultimate target contract
		callData,        // ultimate useful payload
	)

	var initCode []byte
	if !la.isDeployed && nonce == 0 {
		initCode = la.initCode()
	}

	userOperation := &UserOperation{
		Sender:   la.accountAddress,
		Nonce:    big.NewInt(int64(nonce)),
		InitCode: initCode,
		CallData: executeCallData,

		// TODO: implement gas estimates via AA API
		CallGasLimit:         big.NewInt(gasEstimates[0].CallGasLimit),
		VerificationGasLimit: big.NewInt(gasEstimates[0].VerificationGasLimit),
		PreVerificationGas:   big.NewInt(gasEstimates[0].PreVerificationGas),
		MaxFeePerGas:         big.NewInt(gasEstimates[0].MaxFeePerGas),
		MaxPriorityFeePerGas: big.NewInt(gasEstimates[0].MaxPriorityFeePerGas),

		// TODO: implement paymaster support
		Paymaster:                     ethcmn.Address{},
		PaymasterVerificationGasLimit: big.NewInt(300000),
		PaymasterPostOpGasLimit:       ethcmn.Big0,
		PaymasterData:                 []byte{},
	}

	return userOperation, nil
}
