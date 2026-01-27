package payload

import (
	"context"
	"time"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	comettypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
)

var _ TxProvider = &TxnsReplayStressProvider{}

const (
	DEFAULT_TIMEOUT = 1 * time.Second
)

type TxnsReplayStressProvider struct {
	chain.Client
	logger log.Logger

	// Channels for receiving data from sniffer
	blocksCh <-chan *comettypes.Block
	errCh    <-chan error
	doneCh   <-chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTxnsReplayStressProvider(
	client chain.Client,
	blocksCh <-chan *comettypes.Block,
	errCh <-chan error,
	doneCh <-chan struct{},
) (TxProvider, error) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &TxnsReplayStressProvider{
		Client: client,
		logger: log.WithFields(log.Fields{
			"provider": "state-replay-stress",
		}),
		blocksCh: blocksCh,
		errCh:    errCh,
		doneCh:   doneCh,
		ctx:      ctx,
		cancel:   cancel,
	}

	return p, nil
}

type txnsReplayTx struct {
	txBytes []byte
}

func (tx *txnsReplayTx) Bytes() []byte {
	return tx.txBytes
}

func (tx *txnsReplayTx) From() chain.Account {
	return chain.Account{}
}

func (tx *txnsReplayTx) Msgs() []sdk.Msg {
	return nil
}

func (tx *txnsReplayTx) WithBytes(bytes []byte) Tx {
	return &txnsReplayTx{
		txBytes: bytes,
	}
}

func (tx *txnsReplayTx) Provider() TxProvider {
	return nil
}

func (tx *txnsReplayTx) WithProvider(provider TxProvider) Tx {
	return tx
}

// FromIdx returns 0 as state replay transactions don't have account indices
func (tx *txnsReplayTx) FromIdx() int {
	return 0
}

// TxIdx returns 0 as state replay transactions don't have sequence indices
func (tx *txnsReplayTx) TxIdx() int {
	return 0
}

func (p *TxnsReplayStressProvider) Name() string {
	return "state-replay-stress"
}

func (p *TxnsReplayStressProvider) GenerateTx(req TxRequest) (Tx, error) {
	// This method is not used in block-based replay mode
	return nil, errors.New("GenerateTx not supported in block-based replay mode - use GetNextBlockTxs() instead")
}

// GetNextBlockTxs returns all transactions from the next available block
func (p *TxnsReplayStressProvider) GetNextBlockTxs() ([]Tx, error) {
	select {
	case block := <-p.blocksCh:
		if block == nil {
			return nil, errors.New("no more blocks available - processing completed")
		}

		var blockTxs []Tx
		for _, txBytes := range block.Txs {
			if len(txBytes) > 0 {
				blockTxs = append(blockTxs, &txnsReplayTx{txBytes: txBytes})
			}
		}

		p.logger.WithFields(log.Fields{
			"block_height": block.Height,
			"block_txs":    len(blockTxs),
		}).Info("got block transactions")

		return blockTxs, nil

	case err := <-p.errCh:
		return nil, errors.Wrap(err, "sniffer error received")

	case <-p.doneCh:
		return nil, errors.New("no more blocks available - processing completed")

	case <-p.ctx.Done():
		return nil, errors.New("transaction generation cancelled")

	case <-time.After(DEFAULT_TIMEOUT):
		return nil, errors.New("timeout waiting for block from sniffer")
	}
}

func (p *TxnsReplayStressProvider) GenerateInitialTx(req TxRequest) (Tx, error) {
	return nil, nil
}

func (p *TxnsReplayStressProvider) BuildAndSignTx(client chain.Client, unsignedTx Tx) (Tx, error) {
	return unsignedTx.WithBytes(unsignedTx.Bytes()), nil
}

// Stop cancels processing and cleans up resources
func (p *TxnsReplayStressProvider) Stop() {
	p.cancel()
}
