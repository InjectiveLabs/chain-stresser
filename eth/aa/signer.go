package aa

import (
	"fmt"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Signer interface {
	SignUserOperation(
		userOperation *UserOperation,
	) (*PackedUserOperation, ethcmn.Hash, error)
}

type signer struct {
	entryPointAddress ethcmn.Address
	chainID           *big.Int
	personalSignFn    func(address ethcmn.Address, data []byte) ([]byte, error)
	signerAddress     ethcmn.Address
}

func NewSigner(
	entryPointAddress ethcmn.Address,
	chainID *big.Int,
	personalSignFn func(address ethcmn.Address, data []byte) ([]byte, error),
	signerAddress ethcmn.Address,
) Signer {
	return &signer{
		entryPointAddress: entryPointAddress,
		chainID:           chainID,
		personalSignFn:    personalSignFn,
		signerAddress:     signerAddress,
	}
}

func (s *signer) SignUserOperation(
	userOperation *UserOperation,
) (*PackedUserOperation, ethcmn.Hash, error) {
	packedUO := userOperation.Pack()
	opHash := packedUO.GetUserOpHash(s.entryPointAddress, s.chainID)

	// careful: not opHash.Hex()
	textToSign := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", 32, opHash.Bytes()[:32])

	uoSig, err := s.personalSignFn(s.signerAddress, []byte(textToSign))
	if err != nil {
		err = errors.Wrap(err, "failed to sign user operation with personal signer")
		return packedUO, ethcmn.Hash{}, err
	}

	uoSig = ecSigToRPC(uoSig)

	// hol' up cowboy, BaseLightAccount says you need to ABI-prepend the signature type here
	uoSigWithType := []byte{byte(LightAccountSignatureTypeEoa)}
	uoSigWithType = append(uoSigWithType, uoSig...)

	packedUO.Signature = uoSigWithType

	return packedUO, opHash, nil
}

type LightAccountSignatureType byte

const (
	LightAccountSignatureTypeEoa              LightAccountSignatureType = 0
	LightAccountSignatureTypeContract         LightAccountSignatureType = 1
	LightAccountSignatureTypeContractWithAddr LightAccountSignatureType = 2
)

func ecSigToRPC(sig []byte) []byte {
	sig[64] += 27
	return sig
}
