package replay

import (
	"context"
	"fmt"
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
)

type Sniffer struct {
	cfg       *TxReplayConfig
	rpcClient *rpchttp.HTTP
	logger    log.Logger

	blocksCh chan *comettypes.Block
	errCh    chan error
	doneCh   chan struct{}

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

	c, err := rpchttp.New(s.cfg.CometRPC)
	if err != nil {
		return nil, fmt.Errorf("sniffer error: can't create rpc client: %w", err)
	}
	s.rpcClient = c

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

				res, err := s.rpcClient.Block(s.ctx, &height)
				if err != nil {
					return err
				}

				select {
				case s.blocksCh <- res.Block:
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
