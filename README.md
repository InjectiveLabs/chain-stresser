## chain-stresser

```
> chain-stresser run -h

Usage: chain-stresser run [OPTIONS] [NUM_SENDERS]

Runs the stress test by posting orders

Arguments:
  NUM_SENDERS             Amount of parallel sender jobs. Gets key from the corresponding STRESSER_ACCOUNT_MNEMONIC_% (default 1)

Options:
      --cosmos-chain-id   Specify Chain ID of the Cosmos network. (env $STRESSER_COSMOS_CHAIN_ID) (default "injective-1")
      --cosmos-grpc       Cosmos GRPC querying endpoint (env $STRESSER_COSMOS_GRPC) (default "tcp://localhost:9900")
      --tendermint-rpc    Tendermint RPC endpoint (env $STRESSER_TENDERMINT_RPC) (default "http://localhost:26657")
  -K, --key-name          Keyring key name to use. Specify 'env' to use one loaded from STRESSER_ACCOUNT_MNEMONIC (env $STRESSER_KEY_NAME) (default "user1")
  -B, --base-denom        Spot Market base denom (Default: INJ). (default "inj")
  -Q, --quote-denom       Spot Market quote denom (Default: Peggy USDT). (default "peggy0x69efCB62D98f4a6ff5a0b0CFaa4AAbB122e85e08")
  -F, --fee-recipient     Specify trade fee recipient
  -b, --base-decimals     Specify base denom decimals (Defaults for INJ) (default 18)
  -q, --quote-decimals    Specify quote denom decimals (Defaults for USDT) (default 6)
  -D, --backoff-delay     Specify artifical delay for enqueuing the messages. (default "10ms")
```

### Usage

By default stresser uses key from pre-defined keys it has, namely `user1` matching the E2E init script. You can specify your own key:

```
$ export STRESSER_ACCOUNT_MNEMONIC="physical page glare junk return scale subject river token door mirror title"
$ chain-stresser run -K env
```

To run with custom backoff delay (less backoff -> more Msgs in a Tx):
```
$ chain-stresser run -D 3ms
```

To run multiple senders at once in parallel (multi threaded stressing):

```bash
export STRESSER_ACCOUNT_MNEMONIC_1="divide report just assist salad peanut depart song voice decide fringe stumble"
export STRESSER_ACCOUNT_MNEMONIC_2="physical page glare junk return scale subject river token door mirror title"

$ chain-stresser run 2
# the number is arbitrary, can be 100 if you provide 100 keys
INFO[0005] Added inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku as env_1 from ENV
INFO[0005] Added inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r as env_2 from ENV
INFO[0005] Initializing read-only chain client
INFO[0005] Start watching for new Txns from chain
INFO[0005] Existing spot market for INJ/USDT found: 0x17d9b5fb67666df72a5a858eb9b81104b99da760e3036a8243e05532d50e1c7c
INFO[0005] Got sender context for env_1: inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
INFO[0005] Initializing chain client for inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
INFO[0005] Loop sending new orders to inj/peggy0x69efCB62D98f4a6ff5a0b0CFaa4AAbB122e85e08  sender=inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
INFO[0005] Sending first Msg in a single Tx with await   sender=inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
INFO[0005] Got sender context for env_2: inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
INFO[0005] Initializing chain client for inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
INFO[0005] Loop sending new orders to inj/peggy0x69efCB62D98f4a6ff5a0b0CFaa4AAbB122e85e08  sender=inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
INFO[0005] Sending first Msg in a single Tx with await   sender=inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
INFO[0006] Sent and confirmed first Tx                   hash=88148142E11E090082D420C57AA873E41EB894BFC8E50FD6774DADF9386AFFCA sender=inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku
INFO[0006] Sent and confirmed first Tx                   hash=937CC4F6630AD9074DF665D32878C0D7C92626BF352853055CE3268433085A42 sender=inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
```

Note that when you use multiple accounts, env variables should have a suffix (`_1`, `_2`, etc).
