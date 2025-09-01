# State Replay

State replay enables stress testing with real network transactions. First, clone a network's state locally using `injectived bootstrap-devnet`. Then use the built-in sniffer to capture live transactions from the original network and replay them on your local devnetified chain, where they haven't been executed yet.

## How It Works

**Prerequisite**: 
* A devnetified chain state (cloned from any network using the devnetify feature)
* Create tmp directory under /tmp/chain-stresser for storing sniffed txs.

1. **Clone network state** using devnetify to create a local chain with real network state
2. **Start sniffing** from devnetified height + 1 to maintain correct account sequences  
3. **Capture transactions** from the next N blocks on the original network
4. **Store transactions** in binary format: `[8-byte length][tx data][8-byte length][tx data]...`
5. **Replay raw transactions** on the local cloned chain (since they don't exist there)

## Usage

```bash

chain-stresser tx-state-replay --accounts chain-stresser-deploy/instances/0/accounts.json --sniffer-rpc <remote_rpc_to_sniff_from> --sniffer-start-height <devnetified_height+1> --sniffer-end-height <any_future_height> --replay-file-path /tmp/chain-stresser/txns --sniffer-enabled true --sniffer-append true --accounts-num 10 --transactions 10

```
