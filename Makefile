all:

lint:
	golangci-lint run ./...

install:
	go install ./cmd/...

solidity:
	solc --combined-json abi,bin contracts/solidity/Counter.sol > contracts/solidity/Counter.json
	abigen --combined-json contracts/solidity/Counter.json --pkg contract --type Counter --out contracts/solidity/Counter/Counter.go
	rm contracts/solidity/Counter.json

	solc --combined-json abi,bin contracts/solidity/BenchmarkInternalCall.sol > contracts/solidity/BenchmarkInternalCall.json
	abigen --combined-json contracts/solidity/BenchmarkInternalCall.json --pkg contract --type BenchmarkInternalCall --out contracts/solidity/BenchmarkInternalCall/BenchmarkInternalCall.go
	rm contracts/solidity/BenchmarkInternalCall.json

gen-0:
	chain-stresser generate --accounts-num 1000 --validators 1 --sentries 0 --instances 1 --evm true

val-0-start:
	injectived --home="./chain-stresser-deploy/validators/0" start --json-rpc.api eth,web3,net --trace

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

args = $(foreach a,$($(subst _,-,$1)_args),$(if $(value $a),"$($a)"))
eth-counter-get_args = contract

eth-counter-get:
	etherman -N Counter -S ./contracts/solidity/Counter.sol call $(call args,$@) getCount

cook:
	rsync -r ../chain-stresser cooking:~/go/src/

.PHONY: lint install solidity cook
.PHONY: gen-0 val-0-start val-0-clean
.PHONY: run-bank-send run-eth-send run-eth-call
.PHONY: eth-counter-get
