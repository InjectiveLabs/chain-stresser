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
      --rate-tps float         Rate limit transactions per second. Example: 200 for 200 TPS, 99.5 for fractional rates. 0 = no limit.
      --rate-bytes uint        Rate limit transaction bandwidth in bytes per second. Example: 50000 for 50KB/sec. 0 = no limit.
      --rate-gas uint          Rate limit gas consumption per second. Example: 1000000 for 1M gas/sec. 0 = no limit.
      --rate-burst-size int    Custom burst size (tokens). If >0, overrides default burst calculation. 0 = auto.
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

## Rate Limiting (Optional)

Control the stress test intensity with built-in rate limiting. You can limit by transactions per second, bandwidth, or gas consumption.

**Load Generation Pipeline:**
1. **`--accounts-num`** and **`--transactions`** control the total load generated (pipeline capacity)
2. **Rate limits** act as "nozzles" controlling how fast that load flows through

**Example:** 100 accounts × 50 transactions = 5,000 total transactions → rate limiting controls delivery speed

### Transaction Rate Limiting

```bash
# Limit to 100 transactions per second with 10 accounts sending 50 transactions each
chain-stresser tx-bank-send \
  --rate-tps 100 \
  --accounts ./accounts.json \
  --accounts-num 10 \
  --transactions 50

# Fractional rates (0.5 TPS = one transaction every 2 seconds)
chain-stresser tx-bank-send \
  --rate-tps 0.5 \
  --accounts ./accounts.json \
  --accounts-num 5 \
  --transactions 10

# Very high rate for maximum stress (1000 accounts × 100 tx = 100K total transactions)
chain-stresser tx-bank-send \
  --rate-tps 1000 \
  --accounts ./accounts.json \
  --accounts-num 1000 \
  --transactions 100
```

### Bandwidth Rate Limiting

```bash
# Limit to 50KB/sec transaction bandwidth with moderate load
chain-stresser tx-bank-send \
  --rate-bytes 50000 \
  --accounts ./accounts.json \
  --accounts-num 20 \
  --transactions 100

# 1MB/sec bandwidth limit with high load generation
chain-stresser tx-bank-send \
  --rate-bytes 1048576 \
  --accounts ./accounts.json \
  --accounts-num 500 \
  --transactions 200
```

### Gas Rate Limiting

```bash
# Limit to 1 million gas units per second with moderate load
chain-stresser tx-bank-send \
  --rate-gas 1000000 \
  --accounts ./accounts.json \
  --accounts-num 50 \
  --transactions 100

# Conservative gas limiting for ETH contract deployment capacity testing
chain-stresser tx-eth-deploy \
  --rate-gas 500000 \
  --accounts ./accounts.json \
  --accounts-num 10 \
  --transactions 20
```

### Combined Rate Limiting

Rate limits work together as multiple "nozzles" controlling the transaction flow. **The first limit to be reached becomes the bottleneck** that controls the overall rate:

```bash
# Apply multiple limits simultaneously with moderate load generation
chain-stresser tx-bank-send \
  --rate-tps 50 \
  --rate-bytes 100000 \
  --rate-gas 2000000 \
  --rate-burst-size 100 \
  --accounts ./accounts.json \
  --accounts-num 100 \
  --transactions 50

# Network capacity testing with realistic limits and high load generation
chain-stresser tx-eth-call \
  --rate-tps 200 \
  --rate-gas 5000000 \
  --accounts ./accounts.json \
  --accounts-num 200 \
  --transactions 1000
```

**How multiple limits interact:**
- Each transaction must pass **all three checks**: TPS, bytes, and gas
- If TPS limit allows 100 tx/sec but gas limit only allows 50 tx/sec worth of gas → **gas limit wins** (50 tx/sec)
- If bytes limit allows 10 tx/sec but TPS allows 100 tx/sec → **bytes limit wins** (10 tx/sec)
- The **most restrictive limit** determines the actual throughput and they are all optional (default 0, defines no limit).

**Example:** Bank sends use ~150K gas and ~500 bytes each:
```bash
--rate-tps 100 --rate-bytes 10000 --rate-gas 1000000
# TPS allows:   100 tx/sec
# Bytes allows: 10000÷500 = 20 tx/sec  ← BOTTLENECK  
# Gas allows:   1000000÷150000 = 6.6 tx/sec  ← ACTUAL BOTTLENECK
# Result: ~6.6 tx/sec (gas limit wins)
```

### Burst Configuration (Optional)

```bash
# Auto-calculated burst (default behavior, recommended)
chain-stresser tx-bank-send \
  --rate-tps 100 \
  --accounts ./accounts.json \
  --accounts-num 50 \
  --transactions 100

# Manual burst size (allows 50 transactions at once, then rate limited)
chain-stresser tx-bank-send \
  --rate-tps 100 \
  --rate-burst-size 50 \
  --accounts ./accounts.json \
  --accounts-num 50 \
  --transactions 100

# Large burst for bursty workloads
chain-stresser tx-eth-deploy \
  --rate-gas 1000000 \
  --rate-burst-size 500000 \
  --accounts ./accounts.json \
  --accounts-num 20 \
  --transactions 50
```

**Note:** Rate limiting works for all transaction types**

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

## State Replay Feature 
State replay enables stress testing by replaying real network transactions. Checkout [usage guide](state/README.md).

## License

[Apache-2.0](/LICENSE)
