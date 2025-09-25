package ratelimit

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestExtractGasLimit(t *testing.T) {
	// Setup a basic tx config for testing
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	tests := []struct {
		name        string
		gasLimit    uint64
		expectError bool
		expectedGas uint64
	}{
		{
			name:        "normal gas limit",
			gasLimit:    100000,
			expectError: false,
			expectedGas: 100000,
		},
		{
			name:        "high gas limit",
			gasLimit:    10000000,
			expectError: false,
			expectedGas: 10000000,
		},
		{
			name:        "zero gas limit",
			gasLimit:    0,
			expectError: false,
			expectedGas: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple bank send transaction for testing
			txBuilder := txConfig.NewTxBuilder()

			msg := banktypes.NewMsgSend(
				sdk.AccAddress("sender"),
				sdk.AccAddress("recipient"),
				sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			)

			txBuilder.SetMsgs(msg)
			txBuilder.SetGasLimit(tt.gasLimit)
			txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))))

			// Encode the transaction
			txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
			if err != nil {
				t.Fatalf("Failed to encode transaction: %v", err)
			}

			// Test gas extraction
			extractedGas, err := ExtractGasLimit(txConfig, txBytes)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if extractedGas != tt.expectedGas {
				t.Errorf("Expected gas %d, got %d", tt.expectedGas, extractedGas)
			}
		})
	}
}

func TestExtractGasLimit_InvalidTransaction(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	// Test with invalid transaction bytes
	invalidBytes := []byte("invalid transaction data")

	gas, err := ExtractGasLimit(txConfig, invalidBytes)
	if err == nil {
		t.Errorf("Expected error for invalid transaction bytes")
	}
	if gas != 0 {
		t.Errorf("Expected 0 gas for invalid transaction, got %d", gas)
	}
}

func TestExtractGasLimit_NilTxConfig(t *testing.T) {
	// Test with nil txConfig should not panic and return 0 gas
	validBytes := []byte("some transaction bytes")

	gas, err := ExtractGasLimit(nil, validBytes)
	if err != nil {
		t.Errorf("Expected no error for nil txConfig, got: %v", err)
	}
	if gas != 0 {
		t.Errorf("Expected 0 gas for nil txConfig, got %d", gas)
	}
}
