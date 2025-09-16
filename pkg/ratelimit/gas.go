package ratelimit

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExtractGasLimit extracts gas limit from Cosmos SDK transaction bytes.
// This works for both regular Cosmos transactions and Ethereum transactions
// wrapped in MsgEthereumTx since they both become Cosmos SDK transactions.
func ExtractGasLimit(txConfig client.TxConfig, txBytes []byte) (uint64, error) {
	tx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return 0, err
	}

	// Cast to FeeTx to access gas information
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		// If transaction doesn't implement FeeTx, return 0
		return 0, nil
	}

	return feeTx.GetGas(), nil
}
