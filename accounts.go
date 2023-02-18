package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"

	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	chainhd "github.com/InjectiveLabs/sdk-go/chain/crypto/hd"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/cosmos/cosmos-sdk/client"
	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	log "github.com/xlab/suplog"
)

// readEnv is a special utility that reads `.env` file into actual environment variables
// of the current app, similar to `dotenv` Node package.
func readEnv() {
	if envdata, _ := os.ReadFile(".env"); len(envdata) > 0 {
		s := bufio.NewScanner(bytes.NewReader(envdata))
		for s.Scan() {
			txt := s.Text()
			valIdx := strings.IndexByte(txt, '=')
			if valIdx < 0 {
				continue
			}

			strValue := strings.Trim(txt[valIdx+1:], `"`)
			if err := os.Setenv(txt[:valIdx], strValue); err != nil {
				log.WithField("name", txt[:valIdx]).WithError(err).Warningln("failed to override ENV variable")
			}
		}
	}
}

var CosmosAccounts = []Account{
	{Name: "validator1", Mnemonic: "remember huge castle bottom apology smooth avocado ceiling tent brief detect poem"},
	{Name: "validator2", Mnemonic: "capable dismiss rice income open wage unveil left veteran treat vast brave"},
	{Name: "validator3", Mnemonic: "jealous wrist abstract enter erupt hunt victory interest aim defy camp hair"},
	{Name: "user1", Mnemonic: "divide report just assist salad peanut depart song voice decide fringe stumble"},
	{Name: "user2", Mnemonic: "physical page glare junk return scale subject river token door mirror title"},
}

//nolint:all
func getSigningKeys(accounts ...Account) []cryptotypes.PrivKey {
	privkeys := make([]cryptotypes.PrivKey, 0, len(accounts))
	for i := range accounts {
		privkeys = append(privkeys, accounts[i].PrivKey)
	}

	return privkeys
}

func getKeyrings(accounts ...Account) map[string]keyring.Keyring {
	keyrings := make(map[string]keyring.Keyring, len(accounts))
	for i := range accounts {
		keyrings[accounts[i].Name] = accounts[i].Keyring
	}

	return keyrings
}

func getClientContext(chainID, rpcEndpoint string, from ...string) client.Context {
	if len(from) == 0 {
		ctx, err := chainclient.NewClientContext(chainID, "", nil)
		orPanic(err)

		return ctx
	}
	fromName := from[0]

	keyrings := getKeyrings(CosmosAccounts...)
	if _, ok := keyrings[fromName]; !ok {
		orPanic(errors.Errorf("account not found in keyrings: %s", fromName))
	}

	ctx, err := chainclient.NewClientContext(chainID, fromName, keyrings[fromName])
	orPanic(err)

	ctx = ctx.WithNodeURI(rpcEndpoint)
	tmRPC, err := rpchttp.New(rpcEndpoint, "/websocket")
	if err != nil {
		orPanic(err)
	}

	ctx = ctx.WithClient(tmRPC)
	return ctx
}

type Account struct {
	Name     string
	Address  string
	Key      string
	Mnemonic string

	CosmosAccAddress cosmtypes.AccAddress
	CosmosValAddress cosmtypes.ValAddress

	PrivKey cryptotypes.PrivKey
	Keyring keyring.Keyring
}

func (a *Account) Parse() {
	if len(a.Mnemonic) > 0 {
		// derive address and privkey from the provided mnemonic

		algo, err := keyring.NewSigningAlgoFromString("secp256k1", keyring.SigningAlgoList{
			hd.Secp256k1,
		})
		orPanic(err)

		pkBytes, err := algo.Derive()(a.Mnemonic, "", cosmtypes.GetConfig().GetFullBIP44Path())
		orPanic(err)

		cosmosAccPk := &ethsecp256k1.PrivKey{
			Key: pkBytes,
		}

		a.PrivKey = cosmosAccPk
		a.Address = cosmtypes.AccAddress(cosmosAccPk.PubKey().Address().Bytes()).String()

	} else if len(a.Key) > 0 {
		pkBytes, err := hex.DecodeString(a.Key)
		orPanic(err)

		cosmosAccPk := &ethsecp256k1.PrivKey{
			Key: pkBytes,
		}

		a.PrivKey = cosmosAccPk
	}

	accAddress, err := cosmtypes.AccAddressFromBech32(a.Address)
	switch {
	case err == nil:
		// provided a Bech32 address

		if a.PrivKey != nil {
			a.Keyring, err = KeyringForPrivKey(a.Name, a.PrivKey)
			orPanic(err)

			if !bytes.Equal(a.PrivKey.PubKey().Address().Bytes(), accAddress.Bytes()) {
				panic(errors.Errorf("privkey doesn't match address: %s", accAddress.String()))
			}
		}

		a.CosmosAccAddress = accAddress
		a.CosmosValAddress = cosmtypes.ValAddress(accAddress.Bytes())
	case err != nil:
		panic(errors.Wrapf(err, "failed to parse address: %s", a.Address))
	default:
		panic(errors.Errorf("unsupported address: %s", a.Address))
	}
}

// KeyringForPrivKey creates a temporary in-mem keyring for a PrivKey.
// Allows to init Context when the key has been provided in plaintext and parsed.
func KeyringForPrivKey(name string, privKey cryptotypes.PrivKey) (keyring.Keyring, error) {
	kb := keyring.NewInMemory(chainhd.EthSecp256k1Option())
	tmpPhrase := randPhrase(64)
	armored := cosmcrypto.EncryptArmorPrivKey(privKey, tmpPhrase, privKey.Type())
	err := kb.ImportPrivKey(name, armored, tmpPhrase)
	if err != nil {
		err = errors.Wrap(err, "failed to import privkey")
		return nil, err
	}

	return kb, nil
}

func randPhrase(size int) string {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	orPanic(err)

	return string(buf)
}
