.PHONY: build test test-v lint format format-check pre-commit-install ci install-cli js-build js-test clean

BIN := bin/cais

test:
	go test ./... -race -count=1

test-v:
	go test ./... -v -count=1

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/cais

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

js-build:
	npm run js:build

js-test:
	npm run js:test

pre-commit-install:
	pre-commit install

ci: test js-test lint format-check

install-cli:
	go install ./cmd/cais

clean:
	rm -rf bin/ tmp/
