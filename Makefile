all:

lint:
	golangci-lint run ./...

install:
	go install ./cmd/...

solidity:
	solc --optimize --optimize-runs 2000000 --combined-json abi,bin eth/solidity/Counter.sol > eth/solidity/Counter.json
	abigen --combined-json eth/solidity/Counter.json --pkg contract --type Counter --out eth/solidity/Counter/Counter.go
	rm eth/solidity/Counter.json

	solc --optimize --optimize-runs 2000000 --combined-json abi,bin eth/solidity/BenchmarkInternalCall.sol > eth/solidity/BenchmarkInternalCall.json
	abigen --combined-json eth/solidity/BenchmarkInternalCall.json --pkg contract --type BenchmarkInternalCall --out eth/solidity/BenchmarkInternalCall/BenchmarkInternalCall.go
	rm eth/solidity/BenchmarkInternalCall.json

gen-0:
	chain-stresser generate --accounts-num 1000 --validators 1 --sentries 0 --instances 1 --evm true

gen-4:
	chain-stresser generate --accounts-num 1000 --validators 4 --sentries 0 --instances 1 --evm true

gen-4-2:
	chain-stresser generate --accounts-num 1000 --validators 4 --sentries 2 --instances 1 --evm true

val-0-start:
	injectived --home="./chain-stresser-deploy/validators/0" start

val-0-clean:
	injectived --home="./chain-stresser-deploy/validators/0" tendermint unsafe-reset-all

val-1-start:
	injectived --home="./chain-stresser-deploy/validators/1" start

val-1-clean:
	injectived --home="./chain-stresser-deploy/validators/1" tendermint unsafe-reset-all

val-2-start:
	injectived --home="./chain-stresser-deploy/validators/2" start

val-2-clean:
	injectived --home="./chain-stresser-deploy/validators/2" tendermint unsafe-reset-all

val-3-start:
	injectived --home="./chain-stresser-deploy/validators/3" start

val-3-clean:
	injectived --home="./chain-stresser-deploy/validators/3" tendermint unsafe-reset-all

gen-4-native:
	chain-stresser generate --accounts-num 1000 --validators 4 --sentries 0 --instances 1 --evm true --native

compose-up:
	docker compose -f chain-stresser-deploy/docker-compose.yml up -d

compose-down:
	docker compose -f chain-stresser-deploy/docker-compose.yml down

run-bank-send:
	chain-stresser tx-bank-send --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-tokenfactory-burn:
	chain-stresser tx-tokenfactory-burn --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

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

run-exchange-batch-orders:
	chain-stresser tx-exchange-batch-orders --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000 --spot-market-ids 0x1422a13427d5eabd4d8de7907c8340f7e58cb15553a9fd4ad5c90406561886f9 --derivative-market-ids 0x1422a13427d5eabd4d8de7907c8340f7e58cb15553a9fd4ad5c90406561886f9

run-wasm-store-code:
	chain-stresser tx-wasm-store-code --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-wasm-init-contract:
	chain-stresser tx-wasm-init-contract --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-wasm-exec-contract:
	chain-stresser tx-wasm-exec-contract --accounts ./chain-stresser-deploy/instances/0/accounts.json --accounts-num 1000

run-mixed:
	chain-stresser tx-mixed-payload mixed-payload-config.yaml

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
.PHONY: run-bank-send run-tokenfactory-burn run-eth-send run-eth-call
.PHONY: eth-counter-get
