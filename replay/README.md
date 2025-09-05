# State Replay

## Overview

State replay enables stress testing with real network transactions, ported from [chilliass](https://github.com/InjectiveLabs/chilliass). 

**Workflow**: Clone a network's state locally using `injectived bootstrap-devnet`, then use the built-in sniffer to capture live transactions from the original network and replay them on your local devnetified chain where they haven't been executed yet.

## How It Works

**Prerequisite**: 
* A devnetified chain state (cloned from any network using the devnetify feature)

1. **Clone network state** using devnetify to create a local chain with real network state
2. **Start sniffing from correct height --sniffer-start-height** from devnetified height + 1 to maintain correct account sequences  
3. **Replay raw transactions** on the local cloned chain (since they don't exist there)

## Usage

```bash

chain-stresser tx-replay --sniffer-rpc <remote_rpc_to_sniff_from> --sniffer-start-height <devnetified_height+1> --sniffer-end-height <optional_any_future_height>

```
