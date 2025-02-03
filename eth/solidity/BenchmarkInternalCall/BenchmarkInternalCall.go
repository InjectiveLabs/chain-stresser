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

// BenchmarkInternalCallMetaData contains all meta data concerning the BenchmarkInternalCall contract.
var BenchmarkInternalCallMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"iterations\",\"type\":\"uint256\"}],\"name\":\"benchmarkInternalCall\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b5060405161001c9061005a565b604051809103905ff080158015610035573d5f803e3d5ffd5b505f80546001600160a01b0319166001600160a01b0392909216919091179055610066565b60d48061021b83390190565b6101a8806100735f395ff3fe608060405234801561000f575f80fd5b5060043610610029575f3560e01c8063ff04d8061461002d575b5f80fd5b61004061003b366004610106565b610052565b60405190815260200160405180910390f35b5f80805b838110156100ff575f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663f8a8fd6d6040518163ffffffff1660e01b81526004016020604051808303815f875af11580156100c7573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100eb919061011d565b6100f59083610134565b9150600101610056565b5092915050565b5f60208284031215610116575f80fd5b5035919050565b5f6020828403121561012d575f80fd5b5051919050565b8082018082111561016c577f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b9291505056fea26469706673582212200a948c146a4a71a1358978b29421a804a22a49807e9b27f36c7bcd90f1868ab064736f6c63430008170033608060405234801561000f575f80fd5b5060b88061001c5f395ff3fe6080604052348015600e575f80fd5b50600436106026575f3560e01c8063f8a8fd6d14602a575b5f80fd5b60306042565b60405190815260200160405180910390f35b5f7f1440c4dd67b4344ea1905ec0318995133b550f168b4ee959a0da6b503d7d2414602a604051607491815260200190565b60405180910390a150602a9056fea264697066735822122069b5123c635adafc8d2ff9742374307e345a89a93484467e96b0e12d9d93ad3a64736f6c63430008170033",
}

// BenchmarkInternalCallABI is the input ABI used to generate the binding from.
// Deprecated: Use BenchmarkInternalCallMetaData.ABI instead.
var BenchmarkInternalCallABI = BenchmarkInternalCallMetaData.ABI

// BenchmarkInternalCallBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BenchmarkInternalCallMetaData.Bin instead.
var BenchmarkInternalCallBin = BenchmarkInternalCallMetaData.Bin

// DeployBenchmarkInternalCall deploys a new Ethereum contract, binding an instance of BenchmarkInternalCall to it.
func DeployBenchmarkInternalCall(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *BenchmarkInternalCall, error) {
	parsed, err := BenchmarkInternalCallMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BenchmarkInternalCallBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BenchmarkInternalCall{BenchmarkInternalCallCaller: BenchmarkInternalCallCaller{contract: contract}, BenchmarkInternalCallTransactor: BenchmarkInternalCallTransactor{contract: contract}, BenchmarkInternalCallFilterer: BenchmarkInternalCallFilterer{contract: contract}}, nil
}

// BenchmarkInternalCall is an auto generated Go binding around an Ethereum contract.
type BenchmarkInternalCall struct {
	BenchmarkInternalCallCaller     // Read-only binding to the contract
	BenchmarkInternalCallTransactor // Write-only binding to the contract
	BenchmarkInternalCallFilterer   // Log filterer for contract events
}

// BenchmarkInternalCallCaller is an auto generated read-only Go binding around an Ethereum contract.
type BenchmarkInternalCallCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BenchmarkInternalCallTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BenchmarkInternalCallTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BenchmarkInternalCallFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BenchmarkInternalCallFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BenchmarkInternalCallSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BenchmarkInternalCallSession struct {
	Contract     *BenchmarkInternalCall // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// BenchmarkInternalCallCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BenchmarkInternalCallCallerSession struct {
	Contract *BenchmarkInternalCallCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// BenchmarkInternalCallTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BenchmarkInternalCallTransactorSession struct {
	Contract     *BenchmarkInternalCallTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// BenchmarkInternalCallRaw is an auto generated low-level Go binding around an Ethereum contract.
type BenchmarkInternalCallRaw struct {
	Contract *BenchmarkInternalCall // Generic contract binding to access the raw methods on
}

// BenchmarkInternalCallCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BenchmarkInternalCallCallerRaw struct {
	Contract *BenchmarkInternalCallCaller // Generic read-only contract binding to access the raw methods on
}

// BenchmarkInternalCallTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BenchmarkInternalCallTransactorRaw struct {
	Contract *BenchmarkInternalCallTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBenchmarkInternalCall creates a new instance of BenchmarkInternalCall, bound to a specific deployed contract.
func NewBenchmarkInternalCall(address common.Address, backend bind.ContractBackend) (*BenchmarkInternalCall, error) {
	contract, err := bindBenchmarkInternalCall(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BenchmarkInternalCall{BenchmarkInternalCallCaller: BenchmarkInternalCallCaller{contract: contract}, BenchmarkInternalCallTransactor: BenchmarkInternalCallTransactor{contract: contract}, BenchmarkInternalCallFilterer: BenchmarkInternalCallFilterer{contract: contract}}, nil
}

// NewBenchmarkInternalCallCaller creates a new read-only instance of BenchmarkInternalCall, bound to a specific deployed contract.
func NewBenchmarkInternalCallCaller(address common.Address, caller bind.ContractCaller) (*BenchmarkInternalCallCaller, error) {
	contract, err := bindBenchmarkInternalCall(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BenchmarkInternalCallCaller{contract: contract}, nil
}

// NewBenchmarkInternalCallTransactor creates a new write-only instance of BenchmarkInternalCall, bound to a specific deployed contract.
func NewBenchmarkInternalCallTransactor(address common.Address, transactor bind.ContractTransactor) (*BenchmarkInternalCallTransactor, error) {
	contract, err := bindBenchmarkInternalCall(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BenchmarkInternalCallTransactor{contract: contract}, nil
}

// NewBenchmarkInternalCallFilterer creates a new log filterer instance of BenchmarkInternalCall, bound to a specific deployed contract.
func NewBenchmarkInternalCallFilterer(address common.Address, filterer bind.ContractFilterer) (*BenchmarkInternalCallFilterer, error) {
	contract, err := bindBenchmarkInternalCall(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BenchmarkInternalCallFilterer{contract: contract}, nil
}

// bindBenchmarkInternalCall binds a generic wrapper to an already deployed contract.
func bindBenchmarkInternalCall(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BenchmarkInternalCallMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BenchmarkInternalCall *BenchmarkInternalCallRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BenchmarkInternalCall.Contract.BenchmarkInternalCallCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BenchmarkInternalCall *BenchmarkInternalCallRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.BenchmarkInternalCallTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BenchmarkInternalCall *BenchmarkInternalCallRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.BenchmarkInternalCallTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BenchmarkInternalCall *BenchmarkInternalCallCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BenchmarkInternalCall.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BenchmarkInternalCall *BenchmarkInternalCallTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BenchmarkInternalCall *BenchmarkInternalCallTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.contract.Transact(opts, method, params...)
}

// BenchmarkInternalCall is a paid mutator transaction binding the contract method 0xff04d806.
//
// Solidity: function benchmarkInternalCall(uint256 iterations) returns(uint256)
func (_BenchmarkInternalCall *BenchmarkInternalCallTransactor) BenchmarkInternalCall(opts *bind.TransactOpts, iterations *big.Int) (*types.Transaction, error) {
	return _BenchmarkInternalCall.contract.Transact(opts, "benchmarkInternalCall", iterations)
}

// BenchmarkInternalCall is a paid mutator transaction binding the contract method 0xff04d806.
//
// Solidity: function benchmarkInternalCall(uint256 iterations) returns(uint256)
func (_BenchmarkInternalCall *BenchmarkInternalCallSession) BenchmarkInternalCall(iterations *big.Int) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.BenchmarkInternalCall(&_BenchmarkInternalCall.TransactOpts, iterations)
}

// BenchmarkInternalCall is a paid mutator transaction binding the contract method 0xff04d806.
//
// Solidity: function benchmarkInternalCall(uint256 iterations) returns(uint256)
func (_BenchmarkInternalCall *BenchmarkInternalCallTransactorSession) BenchmarkInternalCall(iterations *big.Int) (*types.Transaction, error) {
	return _BenchmarkInternalCall.Contract.BenchmarkInternalCall(&_BenchmarkInternalCall.TransactOpts, iterations)
}

// InnerMetaData contains all meta data concerning the Inner contract.
var InnerMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"TestEvent\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"test\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b5060b88061001c5f395ff3fe6080604052348015600e575f80fd5b50600436106026575f3560e01c8063f8a8fd6d14602a575b5f80fd5b60306042565b60405190815260200160405180910390f35b5f7f1440c4dd67b4344ea1905ec0318995133b550f168b4ee959a0da6b503d7d2414602a604051607491815260200190565b60405180910390a150602a9056fea264697066735822122069b5123c635adafc8d2ff9742374307e345a89a93484467e96b0e12d9d93ad3a64736f6c63430008170033",
}

// InnerABI is the input ABI used to generate the binding from.
// Deprecated: Use InnerMetaData.ABI instead.
var InnerABI = InnerMetaData.ABI

// InnerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use InnerMetaData.Bin instead.
var InnerBin = InnerMetaData.Bin

// DeployInner deploys a new Ethereum contract, binding an instance of Inner to it.
func DeployInner(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Inner, error) {
	parsed, err := InnerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(InnerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Inner{InnerCaller: InnerCaller{contract: contract}, InnerTransactor: InnerTransactor{contract: contract}, InnerFilterer: InnerFilterer{contract: contract}}, nil
}

// Inner is an auto generated Go binding around an Ethereum contract.
type Inner struct {
	InnerCaller     // Read-only binding to the contract
	InnerTransactor // Write-only binding to the contract
	InnerFilterer   // Log filterer for contract events
}

// InnerCaller is an auto generated read-only Go binding around an Ethereum contract.
type InnerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InnerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type InnerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InnerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type InnerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// InnerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type InnerSession struct {
	Contract     *Inner            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// InnerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type InnerCallerSession struct {
	Contract *InnerCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// InnerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type InnerTransactorSession struct {
	Contract     *InnerTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// InnerRaw is an auto generated low-level Go binding around an Ethereum contract.
type InnerRaw struct {
	Contract *Inner // Generic contract binding to access the raw methods on
}

// InnerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type InnerCallerRaw struct {
	Contract *InnerCaller // Generic read-only contract binding to access the raw methods on
}

// InnerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type InnerTransactorRaw struct {
	Contract *InnerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewInner creates a new instance of Inner, bound to a specific deployed contract.
func NewInner(address common.Address, backend bind.ContractBackend) (*Inner, error) {
	contract, err := bindInner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Inner{InnerCaller: InnerCaller{contract: contract}, InnerTransactor: InnerTransactor{contract: contract}, InnerFilterer: InnerFilterer{contract: contract}}, nil
}

// NewInnerCaller creates a new read-only instance of Inner, bound to a specific deployed contract.
func NewInnerCaller(address common.Address, caller bind.ContractCaller) (*InnerCaller, error) {
	contract, err := bindInner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &InnerCaller{contract: contract}, nil
}

// NewInnerTransactor creates a new write-only instance of Inner, bound to a specific deployed contract.
func NewInnerTransactor(address common.Address, transactor bind.ContractTransactor) (*InnerTransactor, error) {
	contract, err := bindInner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &InnerTransactor{contract: contract}, nil
}

// NewInnerFilterer creates a new log filterer instance of Inner, bound to a specific deployed contract.
func NewInnerFilterer(address common.Address, filterer bind.ContractFilterer) (*InnerFilterer, error) {
	contract, err := bindInner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &InnerFilterer{contract: contract}, nil
}

// bindInner binds a generic wrapper to an already deployed contract.
func bindInner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := InnerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Inner *InnerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Inner.Contract.InnerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Inner *InnerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Inner.Contract.InnerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Inner *InnerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Inner.Contract.InnerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Inner *InnerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Inner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Inner *InnerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Inner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Inner *InnerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Inner.Contract.contract.Transact(opts, method, params...)
}

// Test is a paid mutator transaction binding the contract method 0xf8a8fd6d.
//
// Solidity: function test() returns(uint256)
func (_Inner *InnerTransactor) Test(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Inner.contract.Transact(opts, "test")
}

// Test is a paid mutator transaction binding the contract method 0xf8a8fd6d.
//
// Solidity: function test() returns(uint256)
func (_Inner *InnerSession) Test() (*types.Transaction, error) {
	return _Inner.Contract.Test(&_Inner.TransactOpts)
}

// Test is a paid mutator transaction binding the contract method 0xf8a8fd6d.
//
// Solidity: function test() returns(uint256)
func (_Inner *InnerTransactorSession) Test() (*types.Transaction, error) {
	return _Inner.Contract.Test(&_Inner.TransactOpts)
}

// InnerTestEventIterator is returned from FilterTestEvent and is used to iterate over the raw logs and unpacked data for TestEvent events raised by the Inner contract.
type InnerTestEventIterator struct {
	Event *InnerTestEvent // Event containing the contract specifics and raw log

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
func (it *InnerTestEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(InnerTestEvent)
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
		it.Event = new(InnerTestEvent)
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
func (it *InnerTestEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *InnerTestEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// InnerTestEvent represents a TestEvent event raised by the Inner contract.
type InnerTestEvent struct {
	Arg0 *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterTestEvent is a free log retrieval operation binding the contract event 0x1440c4dd67b4344ea1905ec0318995133b550f168b4ee959a0da6b503d7d2414.
//
// Solidity: event TestEvent(uint256 arg0)
func (_Inner *InnerFilterer) FilterTestEvent(opts *bind.FilterOpts) (*InnerTestEventIterator, error) {

	logs, sub, err := _Inner.contract.FilterLogs(opts, "TestEvent")
	if err != nil {
		return nil, err
	}
	return &InnerTestEventIterator{contract: _Inner.contract, event: "TestEvent", logs: logs, sub: sub}, nil
}

// WatchTestEvent is a free log subscription operation binding the contract event 0x1440c4dd67b4344ea1905ec0318995133b550f168b4ee959a0da6b503d7d2414.
//
// Solidity: event TestEvent(uint256 arg0)
func (_Inner *InnerFilterer) WatchTestEvent(opts *bind.WatchOpts, sink chan<- *InnerTestEvent) (event.Subscription, error) {

	logs, sub, err := _Inner.contract.WatchLogs(opts, "TestEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(InnerTestEvent)
				if err := _Inner.contract.UnpackLog(event, "TestEvent", log); err != nil {
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

// ParseTestEvent is a log parse operation binding the contract event 0x1440c4dd67b4344ea1905ec0318995133b550f168b4ee959a0da6b503d7d2414.
//
// Solidity: event TestEvent(uint256 arg0)
func (_Inner *InnerFilterer) ParseTestEvent(log types.Log) (*InnerTestEvent, error) {
	event := new(InnerTestEvent)
	if err := _Inner.contract.UnpackLog(event, "TestEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
