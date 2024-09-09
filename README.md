# Chain Stresser

Our benchamark tool for stress testing the Injective Chain. Configures devnets of any scale and facilitates the execution of transactions from many accounts in parallel.

## Installation

```
git clone https://github.com/InjectiveLabs/chain-stresser.git && cd chain-stresser
git checkout v2
make install
```

## Usage

```
Usage:
  chain-stresser [command]

Available Commands:
  generate     Generates all the config files required to start injectived cluster with state for stress testing.
  tx-bank-send Run stresstest with x/bank.MsgSend transactions.
  tx-eth-send  Run stresstest with eth value send transactions.

Flags:
      --accounts string        Path to a JSON file containing private keys of accounts to use for stress testing. (default "accounts.json")
      --accounts-num int       Number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file. (default 1000)
      --chain-id string        Expected ID of the chain. (default "stressinj-1337")
  -h, --help                   help for chain-stresser
      --min-gas-price string   Minimum gas price to pay for each transaction. (default "1inj")
      --node-addr string       Address of a injectived node RPC to connect to. (default "localhost:26657")
      --transactions int       Number of transactions to allocate for each account. (default 100)
```

## Example

Generate a config for 1 validator and 1000 accounts:

```
chain-stresser generate --accounts-num 1000 --validators 1 --sentries 0 --instances 1
```

Run local validator node with this config:

```
injectived --home="./chain-stresser-deploy/validators/0" start
```

Run a stress test against this node (in separate tab):

```
chain-stresser tx-bank-send --accounts ./chain-stresser-deploy/instances/0/accounts.json
```

## License

[Apache-2.0](/LICENSE)
