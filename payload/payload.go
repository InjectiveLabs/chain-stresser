package payload

import (
	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxRequest struct {
	// Keys is the list of private keys of all accounts stresser iterates.
	// Allows implementation to pick a valid (accessable) receiver account address.
	Keys []chain.Secp256k1PrivateKey

	// From is the account spec that will send the transaction.
	From chain.Account

	// FromIdx is the index of the key that will send the transaction.
	FromIdx int

	// TxIdx is the index of the transaction from the same account.
	TxIdx int
}

type TxProvider interface {
	Name() string

	GenerateInitialTx(
		req TxRequest,
	) (Tx, error)

	GenerateTx(
		req TxRequest,
	) (Tx, error)

	BuildAndSignTx(
		client chain.Client,
		unsignedTx Tx,
	) (signedTx Tx, err error)
}

type QueuedTx interface {
	FromIdx() int
	TxIdx() int
}

type Tx interface {
	QueuedTx

	From() chain.Account
	Msgs() []sdk.Msg
	Bytes() []byte
}

var _ Tx = (*baseTx)(nil)

type baseTx struct {
	from    chain.Account
	msgs    []sdk.Msg
	txBytes []byte

	// queue attributes

	fromIdx int
	txIdx   int
}

func (t *baseTx) From() chain.Account {
	return t.from
}

func (t *baseTx) Msgs() []sdk.Msg {
	return t.msgs
}

func (t *baseTx) Bytes() []byte {
	return t.txBytes
}

func (t *baseTx) FromIdx() int {
	return t.fromIdx
}

func (t *baseTx) TxIdx() int {
	return t.txIdx
}
