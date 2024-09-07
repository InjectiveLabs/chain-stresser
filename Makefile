all:

lint:
	golangci-lint run ./...

install:
	go install ./cmd/...

gen-0:
	chain-stresser generate --validators 1 --sentries 0 --instances 1 --evm true

val-0-start:
	injectived --home="./chain-stresser-deploy/validators/0" start

val-0-clean:
	injectived --home="./chain-stresser-deploy/validators/0" tendermint unsafe-reset-all

run-bank-send:
	chain-stresser tx-bank-send --accounts ./chain-stresser-deploy/instances/0/accounts.json

cook:
	rsync -r ../chain-stresser cooking:~/go/src/

.PHONY: lint install cook gen-0 val-0-start val-0-clean run-bank-send
