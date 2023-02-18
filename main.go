package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	cli "github.com/jawher/mow.cli"
	"github.com/xlab/closer"
	"github.com/xlab/pace"
	log "github.com/xlab/suplog"

	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	chaintypes "github.com/InjectiveLabs/sdk-go/chain/types"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var app = cli.App("chain-stresser", "A chain txn throughput testing and INJ wasting utility")

func main() {
	readEnv()

	app.Before = func() {
		config := cosmtypes.GetConfig()
		chaintypes.SetBech32Prefixes(config)
		chaintypes.SetBip44CoinType(config)

		var loadedEnvMnemonic bool
		if mnemonic := os.Getenv("STRESSER_ACCOUNT_MNEMONIC"); len(mnemonic) > 0 {
			loadedEnvMnemonic = true
			CosmosAccounts = append(CosmosAccounts, Account{
				Name: "env", Mnemonic: mnemonic,
			})
		}

		for idx := range CosmosAccounts {
			account := CosmosAccounts[idx]
			account.Parse()
			CosmosAccounts[idx] = account

			if loadedEnvMnemonic {
				log.Infof("Loaded %s mnemonic from ENV", account.CosmosAccAddress.String())
			}
		}
	}

	app.Command("run", "Runs the stress test by posting orders", runStressing)

	if err := app.Run(os.Args); err != nil {
		log.Errorln(err)
	}
}

func runStressing(c *cli.Cmd) {
	var (
		appLogLevel   *string
		cosmosChainID *string
		cosmosGRPC    *string
		tendermintRPC *string
		keyName       *string
	)

	initGlobalOptions(&appLogLevel)

	initCosmosOptions(
		c,
		&cosmosChainID,
		&cosmosGRPC,
		&tendermintRPC,
		&keyName,
	)

	baseDenom := c.StringOpt("B base-denom", "inj", "Spot Market base denom (Default: INJ).")
	quoteDenom := c.StringOpt("Q quote-denom", "peggy0x69efCB62D98f4a6ff5a0b0CFaa4AAbB122e85e08", "Spot Market quote denom (Default: Peggy USDT).")

	feeRecipient := c.StringOpt("F fee-recipient", "", "Specify trade fee recipient")

	baseDenomDecimals := c.IntOpt("b base-decimals", 18, "Specify base denom decimals (Defaults for INJ)")
	quoteDenomDecimals := c.IntOpt("q quote-decimals", 6, "Specify quote denom decimals (Defaults for USDT)")

	backoffDelay := c.StringOpt("D backoff-delay", "10ms", "Specify artificial delay for enqueuing the messages.")

	numSenders := c.IntArg("NUM_SENDERS", 1, "Amount of parallel sender jobs. Gets key from the corresponding STRESSER_ACCOUNT_MNEMONIC_%")

	c.Spec = "[OPTIONS] [NUM_SENDERS]"

	c.Before = func() {
		log.DefaultLogger.SetLevel(logLevel(*appLogLevel))

		var addedSenders []int
		if *numSenders > 1 {
			for i := 0; i < *numSenders; i++ {
				varName := fmt.Sprintf("STRESSER_ACCOUNT_MNEMONIC_%d", i+1)
				if mnemonic := os.Getenv(varName); len(mnemonic) > 0 {
					CosmosAccounts = append(CosmosAccounts, Account{
						Name: fmt.Sprintf("env_%d", i+1), Mnemonic: mnemonic,
					})

					addedSenders = append(addedSenders, len(CosmosAccounts)-1)
				} else {
					log.Fatalln("Env variable is expected but not defined:", varName)
				}
			}

			for idx := range CosmosAccounts {
				account := CosmosAccounts[idx]
				account.Parse()
				CosmosAccounts[idx] = account
			}

			for _, idx := range addedSenders {
				log.Infof("Added %s as %s from ENV", CosmosAccounts[idx].CosmosAccAddress.String(), CosmosAccounts[idx].Name)
			}
		}
	}

	c.Action = func() {
		defer closer.Close()
		closer.Bind(func() {
			log.Warningln("App Exited")
		})

		log.Infoln("Initializing read-only chain client")
		cc, err := chainclient.NewChainClient(getClientContext(*cosmosChainID, *tendermintRPC), *cosmosGRPC)
		orPanic(err)
		closer.Bind(func() {
			cc.Close()
		})

		log.Infoln("Start watching for new Txns from chain")
		go watchNewTxns(*tendermintRPC)

		marketID := exchangetypes.NewSpotMarketID(*baseDenom, *quoteDenom)
		if !checkSpotMarketExists(cc, *baseDenom, *quoteDenom) {
			log.Fatalf("No spot market for %s/%s found", *baseDenom, *quoteDenom)
		}

		if *feeRecipient == "" {
			*feeRecipient = CosmosAccounts[0].CosmosAccAddress.String()
		}

		backoffDelayParsed, err := time.ParseDuration(*backoffDelay)
		orPanic(err)

		sentOrders := pace.New("sent orders", 10*time.Second, pace.DefaultReporter())
		closer.Bind(func() {
			sentOrders.Pause()
			sentOrders.Report(pace.DefaultReporter())
		})

		wg := new(sync.WaitGroup)
		defer wg.Wait()

		for senderIdx := 0; senderIdx < *numSenders; senderIdx++ {
			var accountKeyName string
			if *numSenders == 1 {
				accountKeyName = *keyName
			} else {
				accountKeyName = fmt.Sprintf("env_%d", senderIdx+1)
			}

			senderCtx := getClientContext(*cosmosChainID, *tendermintRPC, accountKeyName)
			log.Infof("Got sender context for %s: %s", accountKeyName, senderCtx.FromAddress.String())

			log.Infof("Initializing chain client for %s", senderCtx.FromAddress.String())
			cc, err := chainclient.NewChainClient(senderCtx, *cosmosGRPC)
			orPanic(err)
			closer.Bind(func() {
				cc.Close()
			})

			wg.Add(1)
			go func() {
				defer wg.Done()

				makeMsg := func(debugMsg bool) cosmtypes.Msg {
					order := &exchangetypes.SpotOrder{
						MarketId:  marketID.Hex(),
						OrderType: exchangetypes.OrderType_BUY,
						OrderInfo: exchangetypes.OrderInfo{
							SubaccountId: defaultSubaccount(senderCtx.FromAddress).Hex(),
							FeeRecipient: *feeRecipient,
							Price:        getPrice(0.001, *baseDenomDecimals, *quoteDenomDecimals),
							Quantity:     amount(0.1),
						},
					}

					msg := &exchangetypes.MsgCreateSpotLimitOrder{
						Sender: senderCtx.FromAddress.String(),
						Order:  *order,
					}

					if debugMsg {
						v, _ := json.MarshalIndent(msg, "", "\t")
						log.Infoln("Sending Msg:", string(v))
					}

					return msg
				}

				firstOne := true
				senderLog := log.WithField("sender", senderCtx.FromAddress.String())
				senderLog.Infof("Loop sending new orders to %s/%s", *baseDenom, *quoteDenom)
				for {
					if firstOne {
						senderLog.Infoln("Sending first Msg in a single Tx with await")

						txResp, err := cc.SyncBroadcastMsg(makeMsg(true))
						orPanic(err)

						if txResp.TxResponse.Code != 0 {
							senderLog.WithField("hash", txResp.TxResponse.TxHash).Warningln("Tx Error")
							senderLog.Fatalf("sending spot market order Tx error: %s", txResp.String())
						}

						senderLog.WithField("hash", txResp.TxResponse.TxHash).Infoln("Sent and confirmed first Tx")
						firstOne = false
					}

					err = cc.QueueBroadcastMsg(makeMsg(false))
					if err != nil {
						senderLog.WithError(err).Errorln("failed to enqueue Msg")
						return
					}

					sentOrders.Step(1)

					time.Sleep(backoffDelayParsed)
				}
			}()
		}
	}
}

func checkSpotMarketExists(cc chainclient.ChainClient, baseDenom, quoteDenom string) bool {
	ctx, cancelFn := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFn()

	exchangeClient := exchangetypes.NewQueryClient(cc.QueryClient())

	resp, err := exchangeClient.SpotMarket(ctx, &exchangetypes.QuerySpotMarketRequest{
		MarketId: exchangetypes.NewSpotMarketID(baseDenom, quoteDenom).Hex(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "spot market not found") {
			return false
		}

		orPanic(err)
		return false
	}

	log.Infof("Existing spot market for %s found: %s", resp.Market.Ticker, resp.Market.MarketId)
	return true
}

func watchNewTxns(tmRPC string) {
	rpcClient, err := rpchttp.NewWithTimeout(tmRPC, "/websocket", 10)
	if err != nil {
		log.WithError(err).Fatalln("failed to init rpcClient")
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	closer.Bind(func() {
		cancelFn()
	})

	err = rpcClient.Start()
	orPanic(err)

	var out <-chan ctypes.ResultEvent
	out, err = rpcClient.Subscribe(ctx, "chain-stresser", "tm.event = 'Tx'", 100)
	orPanic(err)

	confirmedTxns := pace.New("confirmed txns", 10*time.Second, pace.DefaultReporter())
	closer.Bind(func() {
		confirmedTxns.Pause()
		confirmedTxns.Report(pace.DefaultReporter())
	})

	for range out {
		confirmedTxns.Step(1)
	}
}
