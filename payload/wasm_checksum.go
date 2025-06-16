package payload

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// WasmChecksum represents a hash of the Wasm bytecode that serves as an ID. Must be generated from this library.
// The length of a checksum must always be ChecksumLen.
type WasmChecksum []byte

func (cs WasmChecksum) String() string {
	return hex.EncodeToString(cs)
}

func (cs WasmChecksum) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(cs))
}

func (cs *WasmChecksum) UnmarshalJSON(input []byte) error {
	var hexString string
	err := json.Unmarshal(input, &hexString)
	if err != nil {
		return err
	}

	data, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(data) != WasmChecksumLen {
		return fmt.Errorf("got wrong number of bytes for wasm checksum")
	}
	*cs = WasmChecksum(data)
	return nil
}

const WasmChecksumLen = 32

// WasmCreateChecksum performs the hashing of Wasm bytes to obtain the CosmWasm checksum.
//
// Only Wasm blobs are allowed as inputs and a magic byte check will be performed
// to avoid accidental misusage.
func WasmCreateChecksum(wasm []byte) (WasmChecksum, error) {
	if len(wasm) == 0 {
		return WasmChecksum{}, errors.Errorf("wasm bytes nil or empty")
	}

	if len(wasm) < 4 {
		return WasmChecksum{}, errors.Errorf("wasm bytes shorter than 4 bytes")
	}

	// magic number for Wasm is "\0asm"
	// See https://webassembly.github.io/spec/core/binary/modules.html#binary-module
	if !bytes.Equal(wasm[:4], []byte("\x00\x61\x73\x6D")) {
		return WasmChecksum{}, fmt.Errorf("wasm bytes do not start with Wasm magic number")
	}

	hash := sha256.Sum256(wasm)
	return WasmChecksum(hash[:]), nil
}
