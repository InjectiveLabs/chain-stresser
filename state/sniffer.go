package state

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/xlab/suplog"
	"github.com/ybbus/jsonrpc/v3"
)

const (
	DEFAULT_MAX_RETRIES = 10
)

type Sniffer struct {
	cfg       *StateConfig
	rpcClient jsonrpc.RPCClient
	txnsFile  io.Writer
	logger    log.Logger
	done      chan struct{}
}

type BlockRequest struct {
	Height string `json:"height"`
}

type BlockResult struct {
	Block Block
}

type Block struct {
	Header struct {
		Height string
	}
	Data struct {
		Txs [][]byte
	}
}

func NewSniffer(cfg *StateConfig, txnsBuf io.Writer) (*Sniffer, error) {
	s := &Sniffer{
		cfg: cfg,
		logger: log.WithFields(log.Fields{
			"provider": "state-sniffer",
		}),
		done: make(chan struct{}),
	}

	s.rpcClient = jsonrpc.NewClient(s.cfg.CometRPC)

	if s.cfg.FilePath != "" {
		if err := s.openFile(); err != nil {
			return nil, fmt.Errorf("sniffer error: can't open txns file for write: %w", err)
		}
	} else {
		s.txnsFile = txnsBuf
	}

	return s, nil
}

func (s *Sniffer) Start() error {
	defer close(s.done)

	isProcessing := false
	retries := 0

	for i := s.cfg.StartHeight; s.cfg.EndHeight == 0 || i <= s.cfg.EndHeight; i++ {
		var block *BlockResult
		err := s.rpcClient.CallFor(context.Background(), &block, "block", BlockRequest{Height: fmt.Sprintf("%d", i)})
		if err != nil {
			if !isProcessing { // immediately encountered error
				return err
			} else { // some error popped up during processing, retry this block
				if retries == DEFAULT_MAX_RETRIES {
					return fmt.Errorf("sniffer error: reached maximum number of retries to obtain block %d: %w", i, err)
				}

				time.Sleep(1 * time.Second)

				if retries == 0 {
					i--
				}

				retries++
			}
		} else {
			isProcessing = true
			retries = 0

			if err := s.processBlock(&block.Block); err != nil {
				return fmt.Errorf("sniffer error: during block processing: %w", err)
			}
		}
	}

	s.logger.Info("sniffing completed successfully")
	return nil
}

func (s *Sniffer) Done() <-chan struct{} {
	return s.done
}

func (s *Sniffer) Close() {
	if closer, ok := s.txnsFile.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			s.logger.WithFields(log.Fields{"err": err}).Error("can't close txns file")
		}
	}
}

func (s *Sniffer) openFile() error {
	openFlag := os.O_TRUNC

	if s.cfg.Append {
		openFlag = os.O_APPEND
	}

	f, err := os.OpenFile(s.cfg.FilePath, openFlag|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	s.txnsFile = f
	return nil
}

func (s *Sniffer) processBlock(block *Block) error {
	if s.txnsFile != nil {
		for _, txn := range block.Data.Txs {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(len(txn)))
			if _, err := s.txnsFile.Write(append(b, txn...)); err != nil { // encoded as little-endian 8 bytes len + data
				s.logger.WithFields(log.Fields{"err": err}).Fatal("can't write to output file")
			}
		}
	}

	s.logger.WithFields(log.Fields{"block_height": block.Header.Height, "num": len(block.Data.Txs)}).Info("sniffed")

	return nil
}
