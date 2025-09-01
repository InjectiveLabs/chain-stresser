package payload

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

var _ TxProvider = &stateReplayStressProvider{}

const (
	MAX_TX_SIZE        = 10 * 1024 * 1024 // 10MB max transaction size
	LENGTH_HEADER_SIZE = 8
	MAX_RETRIES        = 10
)

type stateReplayStressProvider struct {
	chain.Client
	replayFilePath string
	replayFile     *os.File
	fileSize       int64
	fileMutex      sync.Mutex
}

func NewStateReplayStressProvider(client chain.Client, replayFilePath string) (TxProvider, error) {
	p := &stateReplayStressProvider{
		Client:         client,
		replayFilePath: replayFilePath,
	}

	if replayFilePath != "" {
		if err := p.openReplayFile(replayFilePath); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("replay file path is required")
	}

	return p, nil
}

type stateReplayTx struct {
	txBytes []byte
	fromIdx int
	txIdx   int
}

func (tx *stateReplayTx) Bytes() []byte {
	return tx.txBytes
}

func (tx *stateReplayTx) From() chain.Account {
	return chain.Account{}
}

func (tx *stateReplayTx) Msgs() []sdk.Msg {
	return nil
}

func (tx *stateReplayTx) WithBytes(bytes []byte) Tx {
	return &stateReplayTx{
		txBytes: bytes,
		fromIdx: tx.fromIdx,
		txIdx:   tx.txIdx,
	}
}

func (tx *stateReplayTx) FromIdx() int {
	return tx.fromIdx
}

func (tx *stateReplayTx) TxIdx() int {
	return tx.txIdx
}

func (p *stateReplayStressProvider) Name() string {
	return "state-replay-stress"
}

func (p *stateReplayStressProvider) GenerateTx(req TxRequest) (Tx, error) {
	p.fileMutex.Lock()
	defer p.fileMutex.Unlock()

	for range MAX_RETRIES {

		lenBuf := make([]byte, LENGTH_HEADER_SIZE)
		lenN, err := p.replayFile.Read(lenBuf)

		// Handle EOF by cycling back to beginning, this covers edge cases when we want to send more tx than its sniffed.
		if err != nil || lenN < LENGTH_HEADER_SIZE {
			if err := p.seekToBeginning(); err != nil {
				return nil, err
			}
			continue
		}

		txLength := binary.LittleEndian.Uint64(lenBuf)
		if txLength == 0 || txLength > MAX_TX_SIZE {
			return nil, errors.New(fmt.Sprintf("invalid transaction length: %d", txLength))
		}

		txnBuf := make([]byte, txLength)
		txnN, err := p.replayFile.Read(txnBuf)
		if err != nil || uint64(txnN) < txLength {
			if err := p.seekToBeginning(); err != nil {
				return nil, err
			}
			continue
		}

		if len(txnBuf) == 0 {
			return nil, errors.New("read empty transaction bytes")
		}

		return &stateReplayTx{
			txBytes: txnBuf,
			fromIdx: req.FromIdx,
			txIdx:   req.TxIdx,
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("failed to read valid transaction after %d retries, empty blocks sniffed", MAX_RETRIES))
}

func (p *stateReplayStressProvider) GenerateInitialTx(req TxRequest) (Tx, error) {
	return nil, nil
}

func (p *stateReplayStressProvider) BuildAndSignTx(client chain.Client, unsignedTx Tx) (Tx, error) {
	return unsignedTx.WithBytes(unsignedTx.Bytes()), nil
}

func (p *stateReplayStressProvider) openReplayFile(filePath string) error {
	for i := range MAX_RETRIES {
		if stat, err := os.Stat(filePath); err == nil {
			if stat.Size() >= LENGTH_HEADER_SIZE {
				break
			}
		}
		if i == MAX_RETRIES-1 {
			return errors.New(fmt.Sprintf("replay file %s does not exist or its missing TX, sniffed empty blocks", filePath))
		}
		time.Sleep(time.Second)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to open replay file %s", filePath))
	}

	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get replay file stats %s", filePath))
	}

	p.replayFile = f
	p.fileSize = stat.Size()
	return nil
}

func (p *stateReplayStressProvider) seekToBeginning() error {
	if _, seekErr := p.replayFile.Seek(0, 0); seekErr != nil {
		return errors.Wrap(seekErr, "can't seek to beginning of replay file")
	}
	return nil
}
