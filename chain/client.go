package chain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	retry "github.com/avast/retry-go/v4"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cmtrpcjson "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	requestTimeout             = 1 * time.Minute
	confirmTimeout             = 10 * time.Minute
	defaultBroadcastStatusPoll = 1 * time.Second
	awaitWarnThreshold         = 1 * time.Minute
	awaitRebroadcastInterval   = 100 * time.Millisecond
	awaitRebroadcastTimeout    = 10 * time.Second
)

var errRetry = errors.New("retry required")

// TODO: replace with https://github.com/InjectiveLabs/sdk-go/tree/master/client/chain
func NewClient(chainID string, addr string, grpcAddr string) Client {
	rpcHTTPClient, err := cmtrpcjson.DefaultHTTPClient("tcp://" + addr)
	orPanic(err)

	// if you're experiencing
	if transport, ok := rpcHTTPClient.Transport.(*http.Transport); ok {
		transport.MaxIdleConns = 4096
		transport.MaxIdleConnsPerHost = 4096
		transport.IdleConnTimeout = 90 * time.Second
	}

	rpcClient, err := rpchttp.NewWithClient("tcp://"+addr, rpcHTTPClient)
	orPanic(err)

	grpcClient, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	orPanic(err)

	clientCtx := NewContext(chainID, rpcClient, grpcClient)

	return Client{
		clientCtx:   clientCtx,
		nodeAddress: addr,
	}
}

type Client struct {
	clientCtx   client.Context
	nodeAddress string
}

func (c Client) TxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c Client) NumUnconfirmedTxs(ctx context.Context) (int, int, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.mempoolQueryURL(), nil)
	if err != nil {
		return 0, 0, 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return 0, 0, 0, errors.Errorf("num_unconfirmed_txs http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, err
	}

	var payload struct {
		Result struct {
			Count      json.RawMessage `json:"n_txs"`
			Total      json.RawMessage `json:"total"`
			TotalBytes json.RawMessage `json:"total_bytes"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, 0, 0, err
	}

	count, err := parseJSONInt(payload.Result.Count)
	if err != nil {
		return 0, 0, 0, err
	}
	total, err := parseJSONInt(payload.Result.Total)
	if err != nil {
		return 0, 0, 0, err
	}
	totalBytes, err := parseJSONInt64(payload.Result.TotalBytes)
	if err != nil {
		return 0, 0, 0, err
	}

	return count, total, totalBytes, nil
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
	Memo     string
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
		tx.Memo,
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

func (c Client) mempoolQueryURL() string {
	addr := c.nodeAddress
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/") + "/num_unconfirmed_txs"
	}

	return "http://" + strings.TrimRight(addr, "/") + "/num_unconfirmed_txs"
}

func parseJSONInt(value json.RawMessage) (int, error) {
	if len(value) == 0 {
		return 0, errors.New("empty json value")
	}

	var asInt int
	if err := json.Unmarshal(value, &asInt); err == nil {
		return asInt, nil
	}

	var asString string
	if err := json.Unmarshal(value, &asString); err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(asString)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func parseJSONInt64(value json.RawMessage) (int64, error) {
	if len(value) == 0 {
		return 0, errors.New("empty json value")
	}

	var asInt int64
	if err := json.Unmarshal(value, &asInt); err == nil {
		return asInt, nil
	}

	var asString string
	if err := json.Unmarshal(value, &asString); err != nil {
		return 0, err
	}

	parsed, err := strconv.ParseInt(asString, 10, 64)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

type broadcastTxSyncResult struct {
	txHash           string
	alreadyInMempool bool
}

func (c Client) Broadcast(ctx context.Context, encodedTx []byte, await bool) (string, error) {
	logger := log.WithField("module", "broadcast")
	var txHash string
	logAwaitDetails := false

	retryOpts := []retry.Option{
		retry.UntilSucceeded(),
		retry.MaxDelay(10 * time.Second),
		retry.MaxJitter(time.Second),
		retry.DelayType(retry.RandomDelay),
	}
	retryOpts = append(retryOpts, retry.Context(ctx))

	// copy bytes just in case
	encodedTxBody := append([]byte{}, encodedTx...)

	if finalError := retry.Do(
		func() error {
			requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
			defer cancel()

			result, err := c.broadcastTxSyncOnce(requestCtx, encodedTxBody)
			if err != nil {
				if IsMempoolFullError(err) {
					return errRetry
				}

				return retry.Unrecoverable(err)
			}

			txHash = result.txHash
			return nil
		},
		retryOpts...,
	); finalError != nil {
		return txHash, finalError
	}

	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode tx hash as hex")
	}

	if !await {
		return txHash, nil
	}

	t := time.NewTimer(defaultBroadcastStatusPoll)
	awaitStart := time.Now()
	warnedAwait := false
	lastRebroadcast := awaitStart

	timeoutCtx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	tryRebroadcast := func() (string, bool, error) {
		rebroadcastCtx, rebroadcastCancel := context.WithTimeout(timeoutCtx, awaitRebroadcastTimeout)
		result, rebroadcastErr := c.broadcastTxSyncOnce(rebroadcastCtx, encodedTxBody)
		rebroadcastCancel()
		if rebroadcastErr != nil {
			if IsMempoolFullError(rebroadcastErr) {
				if logAwaitDetails {
					logger.WithFields(log.Fields{
						"txHash":  txHash,
						"elapsed": time.Since(awaitStart),
					}).WithError(rebroadcastErr).Warning("⚠️ Await rebroadcast blocked by full mempool")
				}
				return "", false, nil
			}
			if logAwaitDetails {
				logger.WithFields(log.Fields{
					"txHash":  txHash,
					"elapsed": time.Since(awaitStart),
				}).WithError(rebroadcastErr).Warning("⚠️ Await rebroadcast failed")
			}
			return "", false, rebroadcastErr
		}

		if !result.alreadyInMempool && result.txHash != "" {
			//logger.WithFields(log.Fields{
			//	"txHash": result.txHash,
			//}).Info("🔄 Tx reinserted into mempool")
			return result.txHash, true, nil
		}

		if logAwaitDetails {
			logger.WithFields(log.Fields{
				"txHash":             txHash,
				"elapsed":            time.Since(awaitStart),
				"already_in_mempool": result.alreadyInMempool,
				"rebroadcast_hash":   result.txHash,
			}).Info("ℹ️ Await rebroadcast result")
		}
		return "", false, nil
	}

	logTxQuery := func(reason string) {
		queryCtx, queryCancel := context.WithTimeout(timeoutCtx, 5*time.Second)
		defer queryCancel()

		resultTx, err := c.clientCtx.Client.Tx(queryCtx, txHashBytes, false)
		if err != nil {
			logger.WithFields(log.Fields{
				"txHash":  txHash,
				"elapsed": time.Since(awaitStart),
				"reason":  reason,
			}).WithError(err).Warning("⚠️ Tx query failed during await")
			return
		}

		logger.WithFields(log.Fields{
			"txHash":    txHash,
			"elapsed":   time.Since(awaitStart),
			"reason":    reason,
			"height":    resultTx.Height,
			"code":      resultTx.TxResult.Code,
			"codespace": resultTx.TxResult.Codespace,
			"log":       resultTx.TxResult.Log,
		}).Info("ℹ️ Tx query result during await")
	}

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
				if isTxIndexingDisabled(err) {
					return "", errors.Errorf("transaction indexing is disabled on the RPC; cannot await tx %s (disable await or enable tx indexing)", txHash)
				}
				if isTxNotFoundError(err) {
					newHash, updated, rebroadcastErr := tryRebroadcast()
					if rebroadcastErr != nil {
						return "", rebroadcastErr
					}
					if updated {
						txHash = newHash
						txHashBytes, err = hex.DecodeString(txHash)
						if err != nil {
							return "", errors.Wrap(err, "failed to decode tx hash as hex")
						}
					}
					lastRebroadcast = time.Now()
				}

			} else if resultTx.TxResult.Code != 0 {
				res := resultTx.TxResult

				if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
					return "", err
				}

				if err := checkNonce(res.Codespace, res.Code, res.Log); err != nil {
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

			elapsed := time.Since(awaitStart)
			if elapsed >= awaitWarnThreshold && !warnedAwait {
				logger.WithFields(log.Fields{
					"txHash":  txHash,
					"elapsed": elapsed,
				}).Warning("⚠️ Awaiting transaction for over 1 minute; will keep polling and reinsert if dropped")
				warnedAwait = true
				logAwaitDetails = true
				logTxQuery("await_warn_threshold")
			}

			if time.Since(lastRebroadcast) >= awaitRebroadcastInterval {
				newHash, updated, rebroadcastErr := tryRebroadcast()
				if rebroadcastErr != nil {
					return "", rebroadcastErr
				}
				if updated {
					txHash = newHash
					txHashBytes, err = hex.DecodeString(txHash)
					if err != nil {
						return "", errors.Wrap(err, "failed to decode tx hash as hex")
					}
				}

				lastRebroadcast = time.Now()
			}

			t.Reset(defaultBroadcastStatusPoll)
		}
	}
}

func (c Client) broadcastTxSyncOnce(ctx context.Context, encodedTx []byte) (broadcastTxSyncResult, error) {
	res, err := c.clientCtx.Client.BroadcastTxSync(ctx, encodedTx)
	if err != nil {
		errRes := client.CheckCometError(err, encodedTx)
		if isMempoolFull(errRes) || isMempoolFullLog(err.Error()) || (errRes != nil && isMempoolFullLog(errRes.RawLog)) {
			msg := ""
			if errRes != nil && errRes.RawLog != "" {
				msg = errRes.RawLog
			} else {
				msg = err.Error()
			}
			return broadcastTxSyncResult{}, errors.Wrap(err, "mempool is full: "+msg)
		}

		if errRes != nil && errRes.Code != 0 {
			logMsg := errRes.RawLog

			if seqErr := checkSequence(errRes.Codespace, errRes.Code, logMsg); seqErr != nil {
				return broadcastTxSyncResult{}, seqErr
			}

			if nonceErr := checkNonce(errRes.Codespace, errRes.Code, logMsg); nonceErr != nil {
				return broadcastTxSyncResult{}, nonceErr
			}
		}

		if seqErr := checkSequenceMessage(err.Error()); seqErr != nil {
			return broadcastTxSyncResult{}, seqErr
		}

		if nonceErr := checkNonceMessage(err.Error()); nonceErr != nil {
			return broadcastTxSyncResult{}, nonceErr
		}

		if isTxInMempool(errRes) {
			return broadcastTxSyncResult{
				txHash:           errRes.TxHash,
				alreadyInMempool: true,
			}, nil
		}

		return broadcastTxSyncResult{}, errors.WithStack(err)
	}

	txHash := res.Hash.String()

	if res.Code != 0 {
		if isMempoolFullLog(res.Log) {
			return broadcastTxSyncResult{}, errors.Errorf("mempool is full: %s", res.Log)
		}

		if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
			return broadcastTxSyncResult{}, err
		}

		if err := checkNonce(res.Codespace, res.Code, res.Log); err != nil {
			return broadcastTxSyncResult{}, err
		}

		err := errors.Errorf(
			"node returned non-zero code for tx '%s' (code: %d, codespace: %s): %s",
			txHash,
			res.Code,
			res.Codespace,
			res.Log,
		)

		return broadcastTxSyncResult{}, err
	}

	return broadcastTxSyncResult{txHash: txHash}, nil
}

func (c Client) NewWasmQueryClient() wasmtypes.QueryClient {
	return wasmtypes.NewQueryClient(c.clientCtx.GRPCClient)
}

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}

	return errRes.Codespace == cosmoserrors.ErrTxInMempoolCache.Codespace() &&
		errRes.Code == cosmoserrors.ErrTxInMempoolCache.ABCICode()
}

func isMempoolFull(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}

	return errRes.Codespace == cosmoserrors.ErrMempoolIsFull.Codespace() &&
		errRes.Code == cosmoserrors.ErrMempoolIsFull.ABCICode()
}

func isMempoolFullLog(msg string) bool {
	if msg == "" {
		return false
	}

	lower := strings.ToLower(msg)
	if strings.Contains(lower, "mempool is full") || strings.Contains(lower, "mempool full") {
		return true
	}

	return strings.Contains(lower, "lane ") && strings.Contains(lower, " is full")
}

func isTxNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tx not found") || strings.Contains(msg, "transaction not found") || strings.Contains(msg, "not found")
}

func isTxIndexingDisabled(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "transaction indexing is disabled") || strings.Contains(msg, "indexing is disabled")
}

func buildAndSignTx(
	clientCtx client.Context,
	signerKey Secp256k1PrivateKey,
	accNum, accSeq uint64,
	fee sdk.Coins,
	gasLimit uint64,
	memo string,
	msgs ...sdk.Msg,
) (signedTx authsigning.Tx, err error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fee)
	if memo != "" {
		txBuilder.SetMemo(memo)
	}

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
var expectedNonceRegExp = regexp.MustCompile(`invalid nonce; got \d+, expected (\d+)`)

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

func checkSequenceMessage(msg string) error {
	if msg == "" {
		return nil
	}

	lower := strings.ToLower(msg)
	matches := expectedSequenceRegExp.FindStringSubmatch(lower)
	if len(matches) != 2 {
		return nil
	}

	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "can't parse expected sequence number, log mesage received: %s", msg)
	}

	return errors.WithStack(sequenceError{
		message:          msg,
		expectedSequence: expectedSequence,
	})
}

func checkNonce(codespace string, code uint32, log string) error {
	if !isSDKErrorResult(codespace, code, cosmoserrors.ErrInvalidSequence) {
		return nil
	}

	matches := expectedNonceRegExp.FindStringSubmatch(log)
	if len(matches) != 2 {
		return errors.Errorf("cosmos sdk hasn't returned expected nonce number, log mesage received: %s", log)
	}

	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "can't parse expected nonce number, log mesage received: %s", log)
	}

	return errors.WithStack(sequenceError{
		message:          log,
		expectedSequence: expectedSequence,
	})
}

func checkNonceMessage(msg string) error {
	if msg == "" {
		return nil
	}

	lower := strings.ToLower(msg)
	matches := expectedNonceRegExp.FindStringSubmatch(lower)
	if len(matches) != 2 {
		return nil
	}

	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "can't parse expected nonce number, log mesage received: %s", msg)
	}

	return errors.WithStack(sequenceError{
		message:          msg,
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

	return isMempoolFullLog(err.Error())
}
