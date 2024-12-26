all:

lint:
	golangci-lint run ./...

install:
	go install ./cmd/...

solidity:
	solc --combined-json abi,bin eth/solidity/Counter.sol > eth/solidity/Counter.json
	abigen --combined-json eth/solidity/Counter.json --pkg contract --type Counter --out eth/solidity/Counter/Counter.go
	rm eth/solidity/Counter.json

	solc --combined-json abi,bin eth/solidity/BenchmarkInternalCall.sol > eth/solidity/BenchmarkInternalCall.json
	abigen --combined-json eth/solidity/BenchmarkInternalCall.json --pkg contract --type BenchmarkInternalCall --out eth/solidity/BenchmarkInternalCall/BenchmarkInternalCall.go
	rm eth/solidity/BenchmarkInternalCall.json

gen-0:
	chain-stresser generate --accounts-num 1000 --validators 1 --sentries 0 --instances 1 --evm true

val-0-start:
	injectived --home="./chain-stresser-deploy/validators/0" start

val-0-clean:
	injectived --home="./chain-stresser-deploy/validators/0" tendermint unsafe-reset-all

run-bank-send:
	chain-stresser tx-bank-send --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-eth-send:
	chain-stresser tx-eth-send --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-eth-call:
	chain-stresser tx-eth-call --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-eth-deploy:
	chain-stresser tx-eth-deploy --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-eth-internal-call:
	chain-stresser tx-eth-internal-call --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000 --iterations 10000

run-eth-userop:
	chain-stresser tx-eth-userop --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 10

args = $(foreach a,$($(subst _,-,$1)_args),$(if $(value $a),"$($a)"))
eth-counter-get_args = contract

eth-counter-get:
	etherman -N Counter -S ./eth/solidity/Counter.sol call $(call args,$@) getCount

eth-counter-deploy:
	etherman -N Counter -S ./eth/solidity/Counter.sol -P 58aeee3e3848e52689b9edca5fccba193c755b02686e6fc34fd13596e5521ebb deploy 0x00

cook:
	rsync -r ../chain-stresser cooking:~/go/src/

.PHONY: lint install solidity cook
.PHONY: gen-0 val-0-start val-0-clean
.PHONY: run-bank-send run-eth-send run-eth-call
.PHONY: eth-counter-get
