package chain

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"text/template"
	"time"

	"cosmossdk.io/math"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	"github.com/cometbft/cometbft/crypto"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"

	_ "embed"
)

const (
	// DefaultVersion is the default app version used in the genesis file
	DefaultVersion = "v1.13.0"

	// DefaultBondDenom is the default bond denomination
	DefaultBondDenom = "inj"

	// DefaultChainID is the default chain ID used in the genesis file
	// Note that the chain ID must end with a number, to allow EVM chain IDs to be used
	DefaultChainID = "stressinj-1337"
)

type GenesisConfig struct {
	AppVersion  string
	GenesisTime string
	ChainID     string
	BondDenom   string
	EvmEnabled  bool
	ProdLike    bool
}

func NewGenesis(genConfig *GenesisConfig) *Genesis {
	if genConfig == nil {
		genConfig = &GenesisConfig{}
	}

	if len(genConfig.AppVersion) == 0 {
		genConfig.AppVersion = DefaultVersion
	}

	if len(genConfig.BondDenom) == 0 {
		genConfig.BondDenom = DefaultBondDenom
	}

	if len(genConfig.ChainID) == 0 {
		genConfig.ChainID = DefaultChainID
	}

	if len(genConfig.GenesisTime) == 0 {
		genConfig.GenesisTime = time.Now().UTC().Format(time.RFC3339)
	}

	buf := new(bytes.Buffer)

	if genConfig.EvmEnabled {
		tpl := template.Must(template.New("genesis_evm").Parse(string(genesisEvmTplJSON)))
		orPanic(tpl.Execute(buf, genConfig))
	} else if genConfig.ProdLike {
		tpl := template.Must(template.New("genesis_prod").Parse(string(genesisProdTplJSON)))
		orPanic(tpl.Execute(buf, genConfig))
	} else {
		tpl := template.Must(template.New("genesis").Parse(string(genesisTplJSON)))
		orPanic(tpl.Execute(buf, genConfig))
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(buf.Bytes())
	orPanic(err)

	var appState map[string]json.RawMessage
	orPanic(json.Unmarshal(genesisDoc.AppState, &appState))

	clientCtx := NewContext(genConfig.ChainID, nil)
	authState := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
	accountState, err := authtypes.UnpackAccounts(authState.Accounts)
	orPanic(err)

	g := &Genesis{
		clientCtx: clientCtx,

		finalized: false,
		mux:       new(sync.Mutex),

		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState),
	}

	// TODO: add pre-defined accounts
	// g.AddAccount()

	return g
}

type Genesis struct {
	clientCtx client.Context

	finalized bool
	mux       *sync.Mutex

	genesisDoc   *tmtypes.GenesisDoc
	appState     map[string]json.RawMessage
	genutilState *genutiltypes.GenesisState
	authState    authtypes.GenesisState
	accountState authtypes.GenesisAccounts
	bankState    *banktypes.GenesisState
}

func (g Genesis) ChainID() string {
	return g.clientCtx.ChainID
}

func (g *Genesis) AddAccount(address sdk.AccAddress, balances string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.verifyNotFinalized()

	baseAccount := authtypes.NewBaseAccount(address, nil, 0, 0)

	g.accountState = append(
		g.accountState,
		&chaintypes.EthAccount{
			BaseAccount: baseAccount,
			CodeHash:    common.BytesToHash(chaintypes.EmptyCodeHash).Bytes(),
		},
	)

	coins, err := sdk.ParseCoinsNormalized(balances)
	orPanic(err)

	g.bankState.Balances = append(
		g.bankState.Balances,
		banktypes.Balance{
			Address: address.String(),
			Coins:   coins,
		},
	)

	g.bankState.Supply = g.bankState.Supply.Add(coins...)
}

func (g *Genesis) AddValidator(
	tmValPubKey crypto.PubKey,
	stakerPrivateKey Secp256k1PrivateKey,
	stakedBalance string,
) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.verifyNotFinalized()

	amount, err := sdk.ParseCoinNormalized(stakedBalance)
	orPanic(err)

	commission := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
	}

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(tmValPubKey)
	orPanic(err)

	stakerAccAddress := sdk.AccAddress((&ethsecp256k1.PrivKey{
		Key: stakerPrivateKey,
	}).PubKey().Address().Bytes())

	valCodec := g.clientCtx.TxConfig.SigningContext().ValidatorAddressCodec()

	valStr, err := valCodec.BytesToString(sdk.ValAddress(stakerAccAddress))
	orPanic(err)

	msg, err := stakingtypes.NewMsgCreateValidator(
		valStr,
		valPubKey,
		amount,
		stakingtypes.Description{
			Moniker: stakerPrivateKey.AccAddress(),
		},
		commission,
		math.OneInt(),
	)
	orPanic(err)

	signedTx, err := buildAndSignTx(g.clientCtx, stakerPrivateKey, 0, 0, sdk.Coins{}, 200000, msg)
	orPanic(err)

	encodedTx := bytesOrPanic(g.clientCtx.TxConfig.TxJSONEncoder()(signedTx))

	g.genutilState.GenTxs = append(
		g.genutilState.GenTxs,
		encodedTx,
	)
}

func (g *Genesis) Save(homeDir string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.finalized = true

	genutiltypes.SetGenesisStateInAppState(g.clientCtx.Codec, g.appState, g.genutilState)

	var err error
	g.authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(g.accountState))
	orPanic(err)

	g.appState[authtypes.ModuleName] = g.clientCtx.Codec.MustMarshalJSON(&g.authState)
	g.bankState.Balances = banktypes.SanitizeGenesisBalances(g.bankState.Balances)
	g.appState[banktypes.ModuleName] = g.clientCtx.Codec.MustMarshalJSON(g.bankState)
	g.genesisDoc.AppState = bytesOrPanic(json.MarshalIndent(g.appState, "", "\t"))

	orPanic(os.MkdirAll(homeDir+"/config", 0o700))
	orPanic(g.genesisDoc.SaveAs(homeDir + "/config/genesis.json"))
}

func (g *Genesis) verifyNotFinalized() {
	if g.finalized {
		panic("genesis has been already saved, no more operations are allowed")
	}
}

var (
	//go:embed templates/genesis.json.tpl
	genesisTplJSON []byte

	//go:embed templates/genesis.evm.json.tpl
	genesisEvmTplJSON []byte

	//go:embed templates/genesis.prod.json.tpl
	genesisProdTplJSON []byte
)
