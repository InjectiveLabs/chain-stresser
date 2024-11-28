package payload

import (
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	evmtypes "github.com/InjectiveLabs/sdk-go/chain/evm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/InjectiveLabs/chain-stresser/v2/chain"
)

type ethTxBuilderAndSigner struct {
	ethSigner ethtypes.Signer
	feeDenom  string
}

func (s *ethTxBuilderAndSigner) BuildAndSignTx(
	client chain.Client,
	unsignedTx Tx,
) (signedTx Tx, err error) {
	ethTxMsg := unsignedTx.Msgs()[0].(*evmtypes.MsgEthereumTx)
	txHash := s.ethSigner.Hash(ethTxMsg.Raw.Transaction)

	privKey := &ethsecp256k1.PrivKey{
		Key: unsignedTx.From().Key,
	}

	sig, err := ethcrypto.Sign(txHash.Bytes(), privKey.ToECDSA())
	if err != nil {
		err = errors.Wrap(err, "failed to sign tx hash")
		return nil, err
	}

	ethTxMsg.Raw.Transaction, err =
		ethTxMsg.Raw.Transaction.WithSignature(s.ethSigner, sig)
	if err != nil {
		err = errors.Wrap(err, "failed to update tx with signature")
		return nil, err
	}

	ethTxMsg.From = privKey.PubKey().Address().Bytes()

	txBuilder := client.TxConfig().NewTxBuilder()

	ethTxOpt, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		err := errors.New("failed to init NewAnyWithValue for ExtensionOptionsEthereumTx")
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		err := errors.New("txBuilder isn't authtx.ExtensionOptionsTxBuilder")
		return nil, err
	} else {
		builder.SetExtensionOptions(ethTxOpt)
	}

	tx, err := ethTxMsg.BuildTx(builder, s.feeDenom)
	if err != nil {
		err = errors.Wrap(err, "failed to build MsgEthereumTx")
		return nil, err
	}

	signedTx = unsignedTx.WithBytes(client.Encode(tx))
	return signedTx, nil
}
