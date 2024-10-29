package chain

import (
	"context"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

const (
	requestTimeout             = 1 * time.Minute
	confirmTimeout             = 10 * time.Minute
	defaultBroadcastStatusPoll = 1 * time.Second
)

// TODO: replace with https://github.com/InjectiveLabs/sdk-go/tree/master/client/chain
func NewClient(chainID string, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	orPanic(err)

	clientCtx := NewContext(chainID, rpcClient)

	return Client{
		clientCtx: clientCtx,
	}
}

type Client struct {
	clientCtx client.Context
}

func (c Client) TxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}
func (c Client) GetNumberSequence(address string) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	orPanic(err)

	accNum, accSeq, err := c.clientCtx.AccountRetriever.GetAccountNumberSequence(c.clientCtx, addr)
	if err != nil {
		return 0, 0, err
	}

	return accNum, accSeq, nil
}

type Tx struct {
	Msgs     []sdk.Msg
	Fee      sdk.Coins
	GasLimit uint64
}

func (c Client) BuildAndSignTx(
	signerAccount Account,
	tx Tx,
) (signedTx authsigning.Tx, err error) {
	return buildAndSignTx(
		c.clientCtx,
		signerAccount.Key,
		signerAccount.Number,
		signerAccount.Sequence,
		tx.Fee,
		tx.GasLimit,
		tx.Msgs...,
	)
}

func (c Client) SignTx(
	signerAccount Account,
	txBuilder client.TxBuilder,
) (signedTx authsigning.Tx, err error) {
	return signTx(
		c.clientCtx,
		signerAccount.Key,
		signerAccount.Number,
		signerAccount.Sequence,
		txBuilder,
	)
}

func (c Client) Encode(signedTx authsigning.Tx) []byte {
	return bytesOrPanic(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

func (c Client) Broadcast(ctx context.Context, encodedTx []byte, await bool) (string, error) {
	var txHash string
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := c.clientCtx.Client.BroadcastTxSync(requestCtx, encodedTx)
	if err != nil {
		errRes := client.CheckCometError(err, encodedTx)
		if !isTxInMempool(errRes) {
			return "", errors.WithStack(err)
		}

		// tx in mempool - all is ok
		txHash = errRes.TxHash
	} else {
		txHash = res.Hash.String()
		if res.Code != 0 {
			if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
				return txHash, err
			}

			err := errors.Errorf(
				"node returned non-zero code for tx '%s' (code: %d, codespace: %s): %s",
				txHash,
				res.Code,
				res.Codespace,
				res.Log,
			)

			return txHash, err
		}
	}

	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode tx hash as hex")
	}

	if !await {
		return txHash, nil
	}

	t := time.NewTimer(defaultBroadcastStatusPoll)

	timeoutCtx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			err := errors.Errorf("transaction timed out while await until included in the block %s", txHash)
			t.Stop()
			return txHash, err

		case <-t.C:
			resultTx, err := c.clientCtx.Client.Tx(timeoutCtx, txHashBytes, false)
			if err != nil {
				if errRes := client.CheckCometError(err, encodedTx); errRes != nil {
					err := errors.Errorf("got tendermint error: %s", errRes.RawLog)
					return "", err
				}

				t.Reset(defaultBroadcastStatusPoll)
				continue

			} else if resultTx.TxResult.Code != 0 {
				res := resultTx.TxResult

				if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
					return "", err
				}

				err = errors.Errorf(
					"node returned non-zero code for tx '%s' (code: %d, codespace: %s): %s",
					txHash,
					res.Code,
					res.Codespace,
					res.Log,
				)

				return txHash, err
			} else if resultTx.Height > 0 {
				t.Stop()

				return txHash, nil
			}

			t.Reset(defaultBroadcastStatusPoll)
		}
	}
}

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}

	return errRes.Codespace == cosmoserrors.ErrTxInMempoolCache.Codespace() &&
		errRes.Code == cosmoserrors.ErrTxInMempoolCache.ABCICode()
}

func buildAndSignTx(
	clientCtx client.Context,
	signerKey Secp256k1PrivateKey,
	accNum, accSeq uint64,
	fee sdk.Coins,
	gasLimit uint64,
	msgs ...sdk.Msg,
) (signedTx authsigning.Tx, err error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fee)
	if err = txBuilder.SetMsgs(msgs...); err != nil {
		err = errors.Wrap(err, "failed to set Tx messages")
		return nil, err
	}

	return signTx(
		clientCtx,
		signerKey,
		accNum,
		accSeq,
		txBuilder,
	)
}

func signTx(
	clientCtx client.Context,
	signerKey Secp256k1PrivateKey,
	accNum, accSeq uint64,
	txBuilder client.TxBuilder,
) (signedTx authsigning.Tx, err error) {

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}

	sigData := &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}

	privKey := &ethsecp256k1.PrivKey{
		Key: signerKey,
	}

	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     sigData,
		Sequence: accSeq,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		err = errors.Wrap(err, "failed to set Tx signatures")
		return nil, err
	}

	bytesToSign := bytesOrPanic(
		authsigning.GetSignBytesAdapter(
			context.Background(),
			clientCtx.TxConfig.SignModeHandler(),
			signing.SignMode_SIGN_MODE_DIRECT,
			signerData,
			txBuilder.GetTx(),
		),
	)

	sigBytes, err := privKey.Sign(bytesToSign)
	if err != nil {
		err = errors.Wrap(err, "failed to sign Tx")
		return nil, err
	}

	sigData.Signature = sigBytes
	if err := txBuilder.SetSignatures(sig); err != nil {
		err = errors.Wrap(err, "failed to set Tx signatures")
		return nil, err
	}

	signedTx = txBuilder.GetTx()
	return signedTx, nil
}

type sequenceError struct {
	expectedSequence uint64
	message          string
}

func (e sequenceError) Error() string {
	return e.message
}

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

func isSDKErrorResult(codespace string, code uint32, sdkErr *errorsmod.Error) bool {
	return codespace == sdkErr.Codespace() &&
		code == sdkErr.ABCICode()
}

func checkSequence(codespace string, code uint32, log string) error {
	// Cosmos SDK doesn't return expected sequence number as a parameter from RPC call,
	// so we must parse the error message in a hacky way.

	if !isSDKErrorResult(codespace, code, cosmoserrors.ErrWrongSequence) {
		return nil
	}

	matches := expectedSequenceRegExp.FindStringSubmatch(log)
	if len(matches) != 2 {
		return errors.Errorf("cosmos sdk hasn't returned expected sequence number, log mesage received: %s", log)
	}

	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "can't parse expected sequence number, log mesage received: %s", log)
	}

	return errors.WithStack(sequenceError{
		message:          log,
		expectedSequence: expectedSequence,
	})
}

// IsSequenceError checks if error is related to account sequence mismatch, and returns expected account sequence
func IsSequenceError(err error) (uint64, bool) {
	var seqErr sequenceError

	if errors.As(err, &seqErr) {
		return seqErr.expectedSequence, true
	}

	return 0, false
}

func IsMempoolFullError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "mempool is full")
}
