package replay

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	comettypes "github.com/cometbft/cometbft/types"
	log "github.com/xlab/suplog"
)

const (
	DEFAULT_MAX_RETRIES   = 10
	DEFAULT_RETRY_DELAY   = 5 * time.Second
	DEFAULT_MAX_DELAY     = 60 * time.Second
	DEFAULT_BLOCKS_BUFFER = 1000

	BlockFromFileTxNum = 100
)

type Sniffer struct {
	cfg       *TxReplayConfig
	rpcClient *rpchttp.HTTP
	logger    log.Logger

	blocksCh chan *comettypes.Block
	errCh    chan error
	doneCh   chan struct{}
	txsFile  io.Reader
	dumpFile io.Writer

	ctx    context.Context
	cancel context.CancelFunc
}

func NewSniffer(cfg *TxReplayConfig, ctx context.Context) (*Sniffer, error) {
	snifferCtx, cancel := context.WithCancel(ctx)

	s := &Sniffer{
		cfg: cfg,
		logger: log.WithFields(log.Fields{
			"provider": "state-sniffer",
		}),
		blocksCh: make(chan *comettypes.Block, DEFAULT_BLOCKS_BUFFER),
		errCh:    make(chan error, 1),
		doneCh:   make(chan struct{}),
		ctx:      snifferCtx,
		cancel:   cancel,
	}

	if cfg.FromFile == "" {
		c, err := rpchttp.New(s.cfg.CometRPC)
		if err != nil {
			return nil, fmt.Errorf("sniffer error: can't create rpc client: %w", err)
		}
		s.rpcClient = c
	} else {
		if f, err := os.OpenFile(cfg.FromFile, os.O_RDONLY, 0644); err != nil {
			return nil, fmt.Errorf("sniffer error: can't open txns file for reading: %w", err)
		} else {
			s.txsFile = f
		}
	}

	if cfg.ToFile != "" {
		if f, err := os.OpenFile(cfg.ToFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return nil, fmt.Errorf("sniffer error: can't open txns dump file for write: %w", err)
		} else {
			s.dumpFile = f
		}
	}

	return s, nil
}

func (s *Sniffer) Start() {
	s.run()
}

func (s *Sniffer) run() {
	defer func() {
		close(s.blocksCh)
		close(s.errCh)
		close(s.doneCh)
	}()
	// Endless mode, if end height is not set keep sniffing forever.
	isEndlessMode := s.cfg.EndHeight == 0
	height := s.cfg.StartHeight

	s.logger.WithFields(log.Fields{
		"start_height": s.cfg.StartHeight,
		"end_height":   s.cfg.EndHeight,
		"endless_mode": isEndlessMode,
		"comet_rpc":    s.cfg.CometRPC,
	}).Info("Starting block sniffer")

	for {
		if !isEndlessMode && height >= s.cfg.EndHeight {
			s.logger.WithFields(log.Fields{"end_height": s.cfg.EndHeight}).Info("reached end height, stopping")
			break
		}

		select {
		case <-s.ctx.Done():
			s.logger.Info("sniffing cancelled")
			return
		default:
		}

		err := retry.Do(
			func() error {
				select {
				case <-s.ctx.Done():
					return s.ctx.Err()
				default:
				}

				var block *comettypes.Block

				if s.cfg.FromFile != "" { // read BlockFromFileTxNum txs from file and construct virtual block
					block = &comettypes.Block{
						Data: comettypes.Data{
							Txs: make([]comettypes.Tx, 0, BlockFromFileTxNum),
						},
					}
					for i := 0; i < BlockFromFileTxNum; i++ {
						tx, err := s.readTxFromFile()
						if err != nil { // end of file or err
							break
						}
						block.Txs = append(block.Txs, tx)
					}

					if len(block.Txs) == 0 {
						block = nil // signal the end
					}
				} else {
					res, err := s.rpcClient.Block(s.ctx, &height)
					if err != nil {
						return err
					}

					block = res.Block

					if s.cfg.ToFile != "" { // just dump txs to file
						if err := s.processBlock(block); err != nil {
							return err
						}
						return nil
					}
				}

				select {
				case s.blocksCh <- block:
				case <-s.ctx.Done():
					return s.ctx.Err()
				}

				return nil
			},
			retry.Attempts(DEFAULT_MAX_RETRIES),
			retry.Delay(DEFAULT_RETRY_DELAY),
			retry.MaxDelay(DEFAULT_MAX_DELAY),
			retry.DelayType(retry.BackOffDelay),
			retry.Context(s.ctx),
			// Retry all errors - if block doesn't exist after DEFAULT_MAX_DELAY, something else is wrong
			retry.RetryIf(func(err error) bool {
				return true
			}),
		)

		if err != nil {
			select {
			case s.errCh <- fmt.Errorf("sniffer error at height %d: %w", height, err):
			case <-s.ctx.Done():
			}
			return
		}

		height++
	}

	s.logger.Info("sniffing completed successfully")
}

func (s *Sniffer) processBlock(block *comettypes.Block) error {
	if s.dumpFile != nil {
		for _, txn := range block.Txs {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(len(txn)))
			if _, err := s.dumpFile.Write(append(b, txn...)); err != nil { // encoded as little-endian 8 bytes len + data
				s.logger.WithFields(log.Fields{"err": err}).Fatal("can't write to output file")
			}
		}
	}

	s.logger.WithFields(log.Fields{"block_height": block.Header.Height, "num": len(block.Data.Txs)}).Info("sniffed")

	return nil
}

func (s *Sniffer) readTxFromFile() ([]byte, error) {
	var (
		lenN   int
		txnN   int
		len    uint64
		txnBuf []byte
	)

	lenBuf := make([]byte, 8)
	lenN, _ = s.txsFile.Read(lenBuf)

	if lenN > 0 {
		len = binary.LittleEndian.Uint64(lenBuf)

		txnBuf = make([]byte, len)
		txnN, _ = s.txsFile.Read(txnBuf)
	}

	if lenN < 8 || uint64(txnN) < len {
		return nil, fmt.Errorf("sniffer error: can't read tx from file")
	}

	return txnBuf, nil
}

// Exposed channel accessors
func (s *Sniffer) Done() <-chan struct{} {
	return s.doneCh
}

func (s *Sniffer) Blocks() <-chan *comettypes.Block {
	return s.blocksCh
}

func (s *Sniffer) Errors() <-chan error {
	return s.errCh
}

func (s *Sniffer) Stop() {
	s.cancel()
}

func (s *Sniffer) Close() {
	s.Stop()
}
