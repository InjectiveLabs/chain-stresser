// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// LightAccountFactoryMetaData contains all meta data concerning the LightAccountFactory contract.
var LightAccountFactoryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"entryPoint\",\"type\":\"address\",\"internalType\":\"contractIEntryPoint\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"receive\",\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"ACCOUNT_IMPLEMENTATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractLightAccount\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ENTRY_POINT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIEntryPoint\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addStake\",\"inputs\":[{\"name\":\"unstakeDelay\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"createAccount\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"contractLightAccount\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAddress\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unlockStake\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdraw\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"addresspayable\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawStake\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"addresspayable\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AddressInsufficientBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"FailedInnerCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAction\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidEntryPoint\",\"inputs\":[{\"name\":\"entryPoint\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TransferFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAddressNotAllowed\",\"inputs\":[]}]",
}

// LightAccountFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use LightAccountFactoryMetaData.ABI instead.
var LightAccountFactoryABI = LightAccountFactoryMetaData.ABI

// LightAccountFactory is an auto generated Go binding around an Ethereum contract.
type LightAccountFactory struct {
	LightAccountFactoryCaller     // Read-only binding to the contract
	LightAccountFactoryTransactor // Write-only binding to the contract
	LightAccountFactoryFilterer   // Log filterer for contract events
}

// LightAccountFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type LightAccountFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightAccountFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LightAccountFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightAccountFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LightAccountFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightAccountFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LightAccountFactorySession struct {
	Contract     *LightAccountFactory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// LightAccountFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LightAccountFactoryCallerSession struct {
	Contract *LightAccountFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// LightAccountFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LightAccountFactoryTransactorSession struct {
	Contract     *LightAccountFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// LightAccountFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type LightAccountFactoryRaw struct {
	Contract *LightAccountFactory // Generic contract binding to access the raw methods on
}

// LightAccountFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LightAccountFactoryCallerRaw struct {
	Contract *LightAccountFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// LightAccountFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LightAccountFactoryTransactorRaw struct {
	Contract *LightAccountFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLightAccountFactory creates a new instance of LightAccountFactory, bound to a specific deployed contract.
func NewLightAccountFactory(address common.Address, backend bind.ContractBackend) (*LightAccountFactory, error) {
	contract, err := bindLightAccountFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactory{LightAccountFactoryCaller: LightAccountFactoryCaller{contract: contract}, LightAccountFactoryTransactor: LightAccountFactoryTransactor{contract: contract}, LightAccountFactoryFilterer: LightAccountFactoryFilterer{contract: contract}}, nil
}

// NewLightAccountFactoryCaller creates a new read-only instance of LightAccountFactory, bound to a specific deployed contract.
func NewLightAccountFactoryCaller(address common.Address, caller bind.ContractCaller) (*LightAccountFactoryCaller, error) {
	contract, err := bindLightAccountFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactoryCaller{contract: contract}, nil
}

// NewLightAccountFactoryTransactor creates a new write-only instance of LightAccountFactory, bound to a specific deployed contract.
func NewLightAccountFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*LightAccountFactoryTransactor, error) {
	contract, err := bindLightAccountFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactoryTransactor{contract: contract}, nil
}

// NewLightAccountFactoryFilterer creates a new log filterer instance of LightAccountFactory, bound to a specific deployed contract.
func NewLightAccountFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*LightAccountFactoryFilterer, error) {
	contract, err := bindLightAccountFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactoryFilterer{contract: contract}, nil
}

// bindLightAccountFactory binds a generic wrapper to an already deployed contract.
func bindLightAccountFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LightAccountFactoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LightAccountFactory *LightAccountFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LightAccountFactory.Contract.LightAccountFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LightAccountFactory *LightAccountFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.LightAccountFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LightAccountFactory *LightAccountFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.LightAccountFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LightAccountFactory *LightAccountFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LightAccountFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LightAccountFactory *LightAccountFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LightAccountFactory *LightAccountFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.contract.Transact(opts, method, params...)
}

// ACCOUNTIMPLEMENTATION is a free data retrieval call binding the contract method 0x290ab984.
//
// Solidity: function ACCOUNT_IMPLEMENTATION() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCaller) ACCOUNTIMPLEMENTATION(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "ACCOUNT_IMPLEMENTATION")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ACCOUNTIMPLEMENTATION is a free data retrieval call binding the contract method 0x290ab984.
//
// Solidity: function ACCOUNT_IMPLEMENTATION() view returns(address)
func (_LightAccountFactory *LightAccountFactorySession) ACCOUNTIMPLEMENTATION() (common.Address, error) {
	return _LightAccountFactory.Contract.ACCOUNTIMPLEMENTATION(&_LightAccountFactory.CallOpts)
}

// ACCOUNTIMPLEMENTATION is a free data retrieval call binding the contract method 0x290ab984.
//
// Solidity: function ACCOUNT_IMPLEMENTATION() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCallerSession) ACCOUNTIMPLEMENTATION() (common.Address, error) {
	return _LightAccountFactory.Contract.ACCOUNTIMPLEMENTATION(&_LightAccountFactory.CallOpts)
}

// ENTRYPOINT is a free data retrieval call binding the contract method 0x94430fa5.
//
// Solidity: function ENTRY_POINT() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCaller) ENTRYPOINT(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "ENTRY_POINT")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ENTRYPOINT is a free data retrieval call binding the contract method 0x94430fa5.
//
// Solidity: function ENTRY_POINT() view returns(address)
func (_LightAccountFactory *LightAccountFactorySession) ENTRYPOINT() (common.Address, error) {
	return _LightAccountFactory.Contract.ENTRYPOINT(&_LightAccountFactory.CallOpts)
}

// ENTRYPOINT is a free data retrieval call binding the contract method 0x94430fa5.
//
// Solidity: function ENTRY_POINT() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCallerSession) ENTRYPOINT() (common.Address, error) {
	return _LightAccountFactory.Contract.ENTRYPOINT(&_LightAccountFactory.CallOpts)
}

// GetAddress is a free data retrieval call binding the contract method 0x8cb84e18.
//
// Solidity: function getAddress(address owner, uint256 salt) view returns(address)
func (_LightAccountFactory *LightAccountFactoryCaller) GetAddress(opts *bind.CallOpts, owner common.Address, salt *big.Int) (common.Address, error) {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "getAddress", owner, salt)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddress is a free data retrieval call binding the contract method 0x8cb84e18.
//
// Solidity: function getAddress(address owner, uint256 salt) view returns(address)
func (_LightAccountFactory *LightAccountFactorySession) GetAddress(owner common.Address, salt *big.Int) (common.Address, error) {
	return _LightAccountFactory.Contract.GetAddress(&_LightAccountFactory.CallOpts, owner, salt)
}

// GetAddress is a free data retrieval call binding the contract method 0x8cb84e18.
//
// Solidity: function getAddress(address owner, uint256 salt) view returns(address)
func (_LightAccountFactory *LightAccountFactoryCallerSession) GetAddress(owner common.Address, salt *big.Int) (common.Address, error) {
	return _LightAccountFactory.Contract.GetAddress(&_LightAccountFactory.CallOpts, owner, salt)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_LightAccountFactory *LightAccountFactorySession) Owner() (common.Address, error) {
	return _LightAccountFactory.Contract.Owner(&_LightAccountFactory.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCallerSession) Owner() (common.Address, error) {
	return _LightAccountFactory.Contract.Owner(&_LightAccountFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_LightAccountFactory *LightAccountFactorySession) PendingOwner() (common.Address, error) {
	return _LightAccountFactory.Contract.PendingOwner(&_LightAccountFactory.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_LightAccountFactory *LightAccountFactoryCallerSession) PendingOwner() (common.Address, error) {
	return _LightAccountFactory.Contract.PendingOwner(&_LightAccountFactory.CallOpts)
}

// RenounceOwnership is a free data retrieval call binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() view returns()
func (_LightAccountFactory *LightAccountFactoryCaller) RenounceOwnership(opts *bind.CallOpts) error {
	var out []interface{}
	err := _LightAccountFactory.contract.Call(opts, &out, "renounceOwnership")

	if err != nil {
		return err
	}

	return err

}

// RenounceOwnership is a free data retrieval call binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() view returns()
func (_LightAccountFactory *LightAccountFactorySession) RenounceOwnership() error {
	return _LightAccountFactory.Contract.RenounceOwnership(&_LightAccountFactory.CallOpts)
}

// RenounceOwnership is a free data retrieval call binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() view returns()
func (_LightAccountFactory *LightAccountFactoryCallerSession) RenounceOwnership() error {
	return _LightAccountFactory.Contract.RenounceOwnership(&_LightAccountFactory.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_LightAccountFactory *LightAccountFactorySession) AcceptOwnership() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.AcceptOwnership(&_LightAccountFactory.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.AcceptOwnership(&_LightAccountFactory.TransactOpts)
}

// AddStake is a paid mutator transaction binding the contract method 0xfbb1c3d4.
//
// Solidity: function addStake(uint32 unstakeDelay, uint256 amount) payable returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) AddStake(opts *bind.TransactOpts, unstakeDelay uint32, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "addStake", unstakeDelay, amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xfbb1c3d4.
//
// Solidity: function addStake(uint32 unstakeDelay, uint256 amount) payable returns()
func (_LightAccountFactory *LightAccountFactorySession) AddStake(unstakeDelay uint32, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.AddStake(&_LightAccountFactory.TransactOpts, unstakeDelay, amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xfbb1c3d4.
//
// Solidity: function addStake(uint32 unstakeDelay, uint256 amount) payable returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) AddStake(unstakeDelay uint32, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.AddStake(&_LightAccountFactory.TransactOpts, unstakeDelay, amount)
}

// CreateAccount is a paid mutator transaction binding the contract method 0x5fbfb9cf.
//
// Solidity: function createAccount(address owner, uint256 salt) returns(address account)
func (_LightAccountFactory *LightAccountFactoryTransactor) CreateAccount(opts *bind.TransactOpts, owner common.Address, salt *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "createAccount", owner, salt)
}

// CreateAccount is a paid mutator transaction binding the contract method 0x5fbfb9cf.
//
// Solidity: function createAccount(address owner, uint256 salt) returns(address account)
func (_LightAccountFactory *LightAccountFactorySession) CreateAccount(owner common.Address, salt *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.CreateAccount(&_LightAccountFactory.TransactOpts, owner, salt)
}

// CreateAccount is a paid mutator transaction binding the contract method 0x5fbfb9cf.
//
// Solidity: function createAccount(address owner, uint256 salt) returns(address account)
func (_LightAccountFactory *LightAccountFactoryTransactorSession) CreateAccount(owner common.Address, salt *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.CreateAccount(&_LightAccountFactory.TransactOpts, owner, salt)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_LightAccountFactory *LightAccountFactorySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.TransferOwnership(&_LightAccountFactory.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.TransferOwnership(&_LightAccountFactory.TransactOpts, newOwner)
}

// UnlockStake is a paid mutator transaction binding the contract method 0xbb9fe6bf.
//
// Solidity: function unlockStake() returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) UnlockStake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "unlockStake")
}

// UnlockStake is a paid mutator transaction binding the contract method 0xbb9fe6bf.
//
// Solidity: function unlockStake() returns()
func (_LightAccountFactory *LightAccountFactorySession) UnlockStake() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.UnlockStake(&_LightAccountFactory.TransactOpts)
}

// UnlockStake is a paid mutator transaction binding the contract method 0xbb9fe6bf.
//
// Solidity: function unlockStake() returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) UnlockStake() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.UnlockStake(&_LightAccountFactory.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0xd9caed12.
//
// Solidity: function withdraw(address to, address token, uint256 amount) returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) Withdraw(opts *bind.TransactOpts, to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "withdraw", to, token, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xd9caed12.
//
// Solidity: function withdraw(address to, address token, uint256 amount) returns()
func (_LightAccountFactory *LightAccountFactorySession) Withdraw(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.Withdraw(&_LightAccountFactory.TransactOpts, to, token, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xd9caed12.
//
// Solidity: function withdraw(address to, address token, uint256 amount) returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) Withdraw(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.Withdraw(&_LightAccountFactory.TransactOpts, to, token, amount)
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xc23a5cea.
//
// Solidity: function withdrawStake(address to) returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) WithdrawStake(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.contract.Transact(opts, "withdrawStake", to)
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xc23a5cea.
//
// Solidity: function withdrawStake(address to) returns()
func (_LightAccountFactory *LightAccountFactorySession) WithdrawStake(to common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.WithdrawStake(&_LightAccountFactory.TransactOpts, to)
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xc23a5cea.
//
// Solidity: function withdrawStake(address to) returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) WithdrawStake(to common.Address) (*types.Transaction, error) {
	return _LightAccountFactory.Contract.WithdrawStake(&_LightAccountFactory.TransactOpts, to)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_LightAccountFactory *LightAccountFactoryTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightAccountFactory.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_LightAccountFactory *LightAccountFactorySession) Receive() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.Receive(&_LightAccountFactory.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_LightAccountFactory *LightAccountFactoryTransactorSession) Receive() (*types.Transaction, error) {
	return _LightAccountFactory.Contract.Receive(&_LightAccountFactory.TransactOpts)
}

// LightAccountFactoryOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the LightAccountFactory contract.
type LightAccountFactoryOwnershipTransferStartedIterator struct {
	Event *LightAccountFactoryOwnershipTransferStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LightAccountFactoryOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LightAccountFactoryOwnershipTransferStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LightAccountFactoryOwnershipTransferStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LightAccountFactoryOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LightAccountFactoryOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LightAccountFactoryOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the LightAccountFactory contract.
type LightAccountFactoryOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*LightAccountFactoryOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _LightAccountFactory.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactoryOwnershipTransferStartedIterator{contract: _LightAccountFactory.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *LightAccountFactoryOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _LightAccountFactory.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LightAccountFactoryOwnershipTransferStarted)
				if err := _LightAccountFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) ParseOwnershipTransferStarted(log types.Log) (*LightAccountFactoryOwnershipTransferStarted, error) {
	event := new(LightAccountFactoryOwnershipTransferStarted)
	if err := _LightAccountFactory.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LightAccountFactoryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the LightAccountFactory contract.
type LightAccountFactoryOwnershipTransferredIterator struct {
	Event *LightAccountFactoryOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LightAccountFactoryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LightAccountFactoryOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LightAccountFactoryOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LightAccountFactoryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LightAccountFactoryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LightAccountFactoryOwnershipTransferred represents a OwnershipTransferred event raised by the LightAccountFactory contract.
type LightAccountFactoryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*LightAccountFactoryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _LightAccountFactory.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &LightAccountFactoryOwnershipTransferredIterator{contract: _LightAccountFactory.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *LightAccountFactoryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _LightAccountFactory.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LightAccountFactoryOwnershipTransferred)
				if err := _LightAccountFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_LightAccountFactory *LightAccountFactoryFilterer) ParseOwnershipTransferred(log types.Log) (*LightAccountFactoryOwnershipTransferred, error) {
	event := new(LightAccountFactoryOwnershipTransferred)
	if err := _LightAccountFactory.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

