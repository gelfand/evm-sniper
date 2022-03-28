GO ?= latest
GORUN = env GO111MODULE=on go run

yul:
	solc --strict-assembly --optimize contracts/main.yul

# generate:
# 	envsubst < /Users/eugene/Developer/go/evm-sniper/contracts/main_template.yul > contracts/main.yul

compile:
	go run -trimpath ./cmd/compile

build:
	mkdir -p ./build/bin/
	go build ./cmd/compile
	mv compile ./build/bin/

.SILENT: generate
