# Chain Stresser

Our benchamark tool for stress testing the Injective Chain. Configures devnets of any scale and facilitates the execution of transactions from many accounts in parallel.

## Installation

```
git clone https://github.com/InjectiveLabs/chain-stresser.git && cd chain-stresser
make install
```

## Usage

```
Usage:
  chain-stresser [command]

Available Commands:
  generate     Generates all the config files required to start injectived cluster with state for stress testing.
  tx-bank-send Run stresstest with x/bank.MsgSend transactions.
  tx-eth-call  Run stresstest with eth contract call transactions.
  tx-eth-send  Run stresstest with eth value send transactions.

Flags:
      --accounts string        Path to a JSON file containing private keys of accounts to use for stress testing. (default "accounts.json")
      --accounts-num int       Number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file. (default 1000)
      --await                  Await for transaction to be included in a block.
      --chain-id string        Expected ID of the chain. (default "stressinj-1337")
  -h, --help                   help for chain-stresser
      --min-gas-price string   Minimum gas price to pay for each transaction. (default "1inj")
      --node-addr string       Address of a injectived node RPC to connect to. (default "localhost:26657")
      --transactions int       Number of transactions to allocate for each account. (default 100)

Use "chain-stresser [command] --help" for more information about a command.
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

## Querying EVM State

Use one of the benchmarks that deploy a contract and update its state. For example, `tx-eth-call` uses a [Counter.sol](./eth/solidity/Counter.sol) contract. You can access its state after benchmark ends using [etherman](https://github.com/InjectiveLabs/etherman) tool.

```
make eth-counter-get contract=0x000...
```

Or use a full CLI command:

```
etherman -N Counter -S ./eth/solidity/Counter.sol call 0x000... getCount
```

See `etherman --help` for more info.

## License

[Apache-2.0](/LICENSE)
