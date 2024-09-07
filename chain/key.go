package chain

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	crypto_cdc "github.com/InjectiveLabs/sdk-go/chain/crypto/codec"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/hd"
	chainsdk "github.com/InjectiveLabs/sdk-go/client/chain"
)

type Secp256k1PrivateKey []byte

func (key Secp256k1PrivateKey) PubKey() Secp256k1PublicKey {
	privKey := ethsecp256k1.PrivKey{
		Key: key,
	}

	return privKey.PubKey().Bytes()
}

func (key Secp256k1PrivateKey) AccAddress() string {
	privKey := ethsecp256k1.PrivKey{
		Key: key,
	}

	return sdk.AccAddress(privKey.PubKey().Address()).String()
}

func (key Secp256k1PrivateKey) ValAddress() string {
	privKey := ethsecp256k1.PrivKey{
		Key: key,
	}

	return sdk.ValAddress(privKey.PubKey().Address()).String()
}

func (key Secp256k1PrivateKey) Address() sdk.AccAddress {
	privKey := ethsecp256k1.PrivKey{
		Key: key,
	}

	return sdk.AccAddress(privKey.PubKey().Address())
}

type Secp256k1PublicKey []byte

func (key Secp256k1PublicKey) Address() sdk.AccAddress {
	pubKey := ethsecp256k1.PubKey{
		Key: key,
	}

	return sdk.AccAddress(pubKey.Address())
}

func GenerateSecp256k1Key() (Secp256k1PublicKey, Secp256k1PrivateKey) {
	privKey, err := ethsecp256k1.GenerateKey()
	orPanic(err)

	return privKey.PubKey().Bytes(), privKey.Bytes()
}

// TODO: use proper keystore abstraction from sdk-go
func WrapKeysInMemKeystore(
	keys map[string]Secp256k1PrivateKey,
) keyring.Keyring {
	keyringDB := keyring.NewInMemory(
		cryptoCdc,
		hd.EthSecp256k1Option(),
	)

	signatureAlgos, _ := keyringDB.SupportedAlgorithms()
	signatureAlgo, err := keyring.NewSigningAlgoFromString(string(hd.EthSecp256k1Type), signatureAlgos)
	orPanic(err)

	// try generating
	signatureAlgo.Generate()

	for name, key := range keys {
		privKey := &ethsecp256k1.PrivKey{
			Key: key,
		}

		encKey := crypto.EncryptArmorPrivKey(privKey, "dummy", privKey.Type())
		orPanic(keyringDB.ImportPrivKey(name, encKey, "dummy"))
	}

	return keyringDB
}

var cryptoCdc *codec.ProtoCodec

func init() {
	cryptoCdc = initCryptoCodec()
}

func initCryptoCodec() *codec.ProtoCodec {
	registry := chainsdk.NewInterfaceRegistry()
	crypto_cdc.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}
