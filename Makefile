.PHONY: dev build test test-v css css-watch docker clean

BIN := bin/cais
CSS_IN := input.css
CSS_OUT := web/static/css/styles.css

test:
	go test ./... -race -count=1

test-v:
	go test ./... -v -count=1

css:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server

AIR := $(shell command -v $(HOME)/go/bin/air 2>/dev/null || command -v air 2>/dev/null)

dev: css
	$(MAKE) css-watch &
	$(AIR) -c .air.toml

docker:
	docker build -t cais:latest .

clean:
	rm -rf bin/ data/ tmp/