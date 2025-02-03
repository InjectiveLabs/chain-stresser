package aa

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	address, _ = abi.NewType("address", "", nil)
	uint256, _ = abi.NewType("uint256", "", nil)
	bytes32, _ = abi.NewType("bytes32", "", nil)

	// PackedUserOpPrimitives is the primitive ABI types for each UserOperation field.
	PackedUserOpPrimitives = []abi.ArgumentMarshaling{
		{Name: "sender", InternalType: "Sender", Type: "address"},
		{Name: "nonce", InternalType: "Nonce", Type: "uint256"},
		{Name: "initCode", InternalType: "InitCode", Type: "bytes"},
		{Name: "callData", InternalType: "CallData", Type: "bytes"},
		{Name: "accountGasLimits", InternalType: "AccountGasLimits", Type: "bytes32"},
		{Name: "preVerificationGas", InternalType: "PreVerificationGas", Type: "uint256"},
		{Name: "gasFees", InternalType: "GasFees", Type: "bytes32"},
		{Name: "paymasterAndData", InternalType: "PaymasterAndData", Type: "bytes"},
		{Name: "signature", InternalType: "Signature", Type: "bytes"},
	}

	// PackedUserOpType is the ABI type of a PackedUserOperation.
	PackedUserOpType, _ = abi.NewType("tuple", "op", PackedUserOpPrimitives)

	// PackedUserOpArr is the ABI type for an array of PackedUserOperations.
	PackedUserOpArr, _ = abi.NewType("tuple[]", "ops", PackedUserOpPrimitives)
)

// UserOperation represents an EIP-4337 style transaction for a smart contract account.
type UserOperation struct {
	Sender                        ethcmn.Address `json:"sender"                        mapstructure:"sender"                        validate:"required"`
	Nonce                         *big.Int       `json:"nonce"                         mapstructure:"nonce"                         validate:"required"`
	InitCode                      []byte         `json:"initCode"                      mapstructure:"initCode"                      validate:"required"`
	CallData                      []byte         `json:"callData"                      mapstructure:"callData"                      validate:"required"`
	CallGasLimit                  *big.Int       `json:"callGasLimit"                  mapstructure:"callGasLimit"                  validate:"required"`
	VerificationGasLimit          *big.Int       `json:"verificationGasLimit"          mapstructure:"verificationGasLimit"          validate:"required"`
	PreVerificationGas            *big.Int       `json:"preVerificationGas"            mapstructure:"preVerificationGas"            validate:"required"`
	MaxFeePerGas                  *big.Int       `json:"maxFeePerGas"                  mapstructure:"maxFeePerGas"                  validate:"required"`
	MaxPriorityFeePerGas          *big.Int       `json:"maxPriorityFeePerGas"          mapstructure:"maxPriorityFeePerGas"          validate:"required"`
	Paymaster                     ethcmn.Address `json:"paymaster"                     mapstructure:"paymaster"                     validate:"required"`
	PaymasterVerificationGasLimit *big.Int       `json:"paymasterVerificationGasLimit" mapstructure:"paymasterVerificationGasLimit" validate:"required"`
	PaymasterPostOpGasLimit       *big.Int       `json:"paymasterPostOpGasLimit"       mapstructure:"paymasterPostOpGasLimit"       validate:"required"`
	PaymasterData                 []byte         `json:"paymasterData"                 mapstructure:"paymasterData"                 validate:"required"`
	Signature                     []byte         `json:"signature"                     mapstructure:"signature"                     validate:"required"`
}

type PackedUserOperation struct {
	Sender             ethcmn.Address `json:"sender"             mapstructure:"sender"             validate:"required"`
	Nonce              *big.Int       `json:"nonce"              mapstructure:"nonce"              validate:"required"`
	InitCode           []byte         `json:"initCode"           mapstructure:"initCode"           validate:"required"`
	CallData           []byte         `json:"callData"           mapstructure:"callData"           validate:"required"`
	AccountGasLimits   [32]byte       `json:"accountGasLimits"   mapstructure:"accountGasLimits"   validate:"required"`
	PreVerificationGas *big.Int       `json:"preVerificationGas" mapstructure:"preVerificationGas" validate:"required"`
	GasFees            [32]byte       `json:"gasFees"            mapstructure:"gasFees"            validate:"required"`
	PaymasterAndData   []byte         `json:"paymasterAndData"   mapstructure:"paymasterAndData"   validate:"required"`
	Signature          []byte         `json:"signature"          mapstructure:"signature"          validate:"required"`
}

// GetFactory returns the address portion of InitCode if applicable. Otherwise it returns the zero address.
func (op *UserOperation) GetFactory() ethcmn.Address {
	if len(op.InitCode) < ethcmn.AddressLength {
		return ethcmn.HexToAddress("0x")
	}

	return ethcmn.BytesToAddress(op.InitCode[:ethcmn.AddressLength])
}

// GetFactoryData returns the data portion of InitCode if applicable. Otherwise it returns an empty byte
// array.
func (op *UserOperation) GetFactoryData() []byte {
	if len(op.InitCode) < ethcmn.AddressLength {
		return []byte{}
	}

	return op.InitCode[ethcmn.AddressLength:]
}

// GetMaxGasAvailable returns the max amount of gas that can be consumed by this UserOperation.
func (op *UserOperation) GetMaxGasAvailable() *big.Int {
	// TODO: Multiplier logic might change in v0.7
	mul := big.NewInt(1)
	if op.Paymaster != ethcmn.HexToAddress("0x") {
		mul = big.NewInt(3)
	}

	return big.NewInt(0).Add(
		big.NewInt(0).Mul(op.VerificationGasLimit, mul),
		big.NewInt(0).Add(op.PreVerificationGas, op.CallGasLimit),
	)
}

// GetMaxPrefund returns the max amount of wei required to pay for gas fees by either the sender or
// paymaster.
func (op *UserOperation) GetMaxPrefund() *big.Int {
	return big.NewInt(0).Mul(op.GetMaxGasAvailable(), op.MaxFeePerGas)
}

// GetDynamicGasPrice returns the effective gas price paid by the UserOperation given a basefee. If basefee is
// nil, it will assume a value of 0.
func (op *UserOperation) GetDynamicGasPrice(basefee *big.Int) *big.Int {
	bf := basefee
	if bf == nil {
		bf = big.NewInt(0)
	}

	gp := big.NewInt(0).Add(bf, op.MaxPriorityFeePerGas)
	if gp.Cmp(op.MaxFeePerGas) == 1 {
		return op.MaxFeePerGas
	}
	return gp
}

// EncodeForSignature returns a minimal message of the userOp. This can be used to generate a userOpHash.
func (op *PackedUserOperation) EncodeForSignature() []byte {
	args := abi.Arguments{
		{Name: "sender", Type: address},
		{Name: "nonce", Type: uint256},
		{Name: "hashInitCode", Type: bytes32},
		{Name: "hashCallData", Type: bytes32},
		{Name: "accountGasLimits", Type: bytes32},
		{Name: "preVerificationGas", Type: uint256},
		{Name: "gasFees", Type: bytes32},
		{Name: "hashPaymasterAndData", Type: bytes32},
	}

	packed, _ := args.Pack(
		op.Sender,
		op.Nonce,
		crypto.Keccak256Hash(op.InitCode),
		crypto.Keccak256Hash(op.CallData),
		op.AccountGasLimits,
		op.PreVerificationGas,
		op.GasFees,
		crypto.Keccak256Hash(op.PaymasterAndData),
	)

	return packed
}

// Encode returns a standard message of the userOp. This cannot be used to generate a userOpHash.
func (op *UserOperation) Encode() []byte {
	packedUO := op.Pack()

	args := abi.Arguments{
		{Name: "PackedUserOp", Type: PackedUserOpType},
	}

	encoded, _ := args.Pack(&struct {
		Sender             ethcmn.Address
		Nonce              *big.Int
		InitCode           []byte
		CallData           []byte
		AccountGasLimits   [32]byte
		PreVerificationGas *big.Int
		GasFees            [32]byte
		PaymasterAndData   []byte
		Signature          []byte
	}{
		packedUO.Sender,
		packedUO.Nonce,
		packedUO.InitCode,
		packedUO.CallData,
		packedUO.AccountGasLimits,
		packedUO.PreVerificationGas,
		packedUO.GasFees,
		packedUO.PaymasterAndData,
		packedUO.Signature,
	})

	return encoded
}

func (op *UserOperation) Pack() *PackedUserOperation {
	accountGasLimits := packAccountGasLimits(op.VerificationGasLimit, op.CallGasLimit)
	if len(accountGasLimits) != 32 {
		panic("accountGasLimits must be 32 bytes")
	}

	gasFees := packGasFees(op.MaxPriorityFeePerGas, op.MaxFeePerGas)
	if len(gasFees) != 32 {
		panic("gasFees must be 32 bytes")
	}

	paymasterAndData := []byte{}
	if !bytes.Equal(op.Paymaster.Bytes(), ethcmn.Address{}.Bytes()) {
		paymasterAndData = packPaymasterData(
			op.Paymaster,
			op.PaymasterVerificationGasLimit,
			op.PaymasterPostOpGasLimit,
			op.PaymasterData,
		)
	}

	return &PackedUserOperation{
		Sender:             op.Sender,
		Nonce:              op.Nonce,
		CallData:           op.CallData,
		AccountGasLimits:   ethcmn.BytesToHash(accountGasLimits),
		InitCode:           op.InitCode,
		PreVerificationGas: op.PreVerificationGas,
		GasFees:            ethcmn.BytesToHash(gasFees),
		PaymasterAndData:   paymasterAndData,
		Signature:          op.Signature,
	}
}

func packAccountGasLimits(verificationGasLimit *big.Int, callGasLimit *big.Int) []byte {
	// Left pad both values to 16 bytes (128 bits) each
	verificationGasLimitBytes := ethcmn.LeftPadBytes(verificationGasLimit.Bytes(), 16)
	callGasLimitBytes := ethcmn.LeftPadBytes(callGasLimit.Bytes(), 16)

	// Concatenate the padded bytes
	return append(verificationGasLimitBytes, callGasLimitBytes...)
}

func packGasFees(maxPriorityFeePerGas *big.Int, maxFeePerGas *big.Int) []byte {
	// Left pad both values to 16 bytes (128 bits) each
	maxPriorityFeePerGasBytes := ethcmn.LeftPadBytes(maxPriorityFeePerGas.Bytes(), 16)
	maxFeePerGasBytes := ethcmn.LeftPadBytes(maxFeePerGas.Bytes(), 16)

	// Concatenate the padded bytes
	return append(maxPriorityFeePerGasBytes, maxFeePerGasBytes...)
}

func packPaymasterData(paymaster ethcmn.Address, paymasterVerificationGasLimit *big.Int, postOpGasLimit *big.Int, paymasterData []byte) []byte {
	// Left pad gas limits to 16 bytes (128 bits) each
	verificationGasLimitBytes := ethcmn.LeftPadBytes(paymasterVerificationGasLimit.Bytes(), 16)
	postOpGasLimitBytes := ethcmn.LeftPadBytes(postOpGasLimit.Bytes(), 16)

	// Concatenate paymaster address, gas limits and paymaster data
	result := paymaster.Bytes()
	result = append(result, verificationGasLimitBytes...)
	result = append(result, postOpGasLimitBytes...)
	result = append(result, paymasterData...)

	return result
}

// GetUserOpHash returns the hash of the userOp + entryPoint address + chainID.
func (op *PackedUserOperation) GetUserOpHash(entryPoint ethcmn.Address, chainID *big.Int) ethcmn.Hash {
	// Get the hash of the user operation
	opEncodedForSignature := op.EncodeForSignature()

	userOpHash := crypto.Keccak256Hash(opEncodedForSignature)

	args := abi.Arguments{
		{Name: "userOpHash", Type: bytes32},
		{Name: "entryPoint", Type: address},
		{Name: "chainId", Type: uint256},
	}

	packed, err := args.Pack(
		userOpHash,
		entryPoint,
		chainID,
	)
	if err != nil {
		panic(err)
	}

	uoHash := crypto.Keccak256Hash(packed)
	return uoHash
}

// MarshalJSON returns a JSON encoding of the UserOperation.
func (op *UserOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Sender                        string `json:"sender"`
		Nonce                         string `json:"nonce"`
		InitCode                      string `json:"initCode"`
		CallData                      string `json:"callData"`
		CallGasLimit                  string `json:"callGasLimit"`
		VerificationGasLimit          string `json:"verificationGasLimit"`
		PreVerificationGas            string `json:"preVerificationGas"`
		MaxFeePerGas                  string `json:"maxFeePerGas"`
		MaxPriorityFeePerGas          string `json:"maxPriorityFeePerGas"`
		Paymaster                     string `json:"paymaster"`
		PaymasterVerificationGasLimit string `json:"paymasterVerificationGasLimit"`
		PaymasterPostOpGasLimit       string `json:"paymasterPostOpGasLimit"`
		PaymasterData                 string `json:"paymasterData"`
		Signature                     string `json:"signature"`
	}{
		Sender:                        op.Sender.String(),
		Nonce:                         hexutil.EncodeBig(op.Nonce),
		InitCode:                      hexutil.Encode(op.InitCode),
		CallData:                      hexutil.Encode(op.CallData),
		CallGasLimit:                  hexutil.EncodeBig(op.CallGasLimit),
		VerificationGasLimit:          hexutil.EncodeBig(op.VerificationGasLimit),
		PreVerificationGas:            hexutil.EncodeBig(op.PreVerificationGas),
		MaxFeePerGas:                  hexutil.EncodeBig(op.MaxFeePerGas),
		MaxPriorityFeePerGas:          hexutil.EncodeBig(op.MaxPriorityFeePerGas),
		Paymaster:                     op.Paymaster.String(),
		PaymasterVerificationGasLimit: hexutil.EncodeBig(op.PaymasterVerificationGasLimit),
		PaymasterPostOpGasLimit:       hexutil.EncodeBig(op.PaymasterPostOpGasLimit),
		PaymasterData:                 hexutil.Encode(op.PaymasterData),
		Signature:                     hexutil.Encode(op.Signature),
	})
}

func (op *PackedUserOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Sender             string `json:"sender"`
		Nonce              string `json:"nonce"`
		InitCode           string `json:"initCode"`
		CallData           string `json:"callData"`
		AccountGasLimits   string `json:"accountGasLimits"`
		PreVerificationGas string `json:"preVerificationGas"`
		GasFees            string `json:"gasFees"`
		PaymasterAndData   string `json:"paymasterAndData"`
		Signature          string `json:"signature"`
	}{
		Sender:             op.Sender.String(),
		Nonce:              hexutil.EncodeBig(op.Nonce),
		InitCode:           hexutil.Encode(op.InitCode),
		CallData:           hexutil.Encode(op.CallData),
		AccountGasLimits:   hexutil.Encode(op.AccountGasLimits[:]),
		PreVerificationGas: hexutil.EncodeBig(op.PreVerificationGas),
		GasFees:            hexutil.Encode(op.GasFees[:]),
		PaymasterAndData:   hexutil.Encode(op.PaymasterAndData),
		Signature:          hexutil.Encode(op.Signature),
	})
}

// ToMap returns the current UserOperation struct as a map type.
func (op *UserOperation) ToMap() (map[string]any, error) {
	data, err := op.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var opData map[string]any
	if err := json.Unmarshal(data, &opData); err != nil {
		return nil, err
	}
	return opData, nil
}

// ToMap returns the current PackedUserOperation struct as a map type.
func (op *PackedUserOperation) ToMap() (map[string]any, error) {
	data, err := op.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var opData map[string]any
	if err := json.Unmarshal(data, &opData); err != nil {
		return nil, err
	}
	return opData, nil
}

type UserOperationGasEstimates struct {
	CallGasLimit         int64
	VerificationGasLimit int64
	PreVerificationGas   int64
	MaxFeePerGas         int64
	MaxPriorityFeePerGas int64
}
