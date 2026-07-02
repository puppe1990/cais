package cli

const tplInputCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer components {
  .htmx-swapping {
    opacity: 0;
    transition: opacity 150ms ease-out;
  }

  .htmx-settling {
    opacity: 1;
    transition: opacity 150ms ease-in;
  }

  form.htmx-request button[type="submit"] {
    @apply opacity-60 pointer-events-none;
  }

  .htmx-indicator {
    @apply hidden;
  }

  .htmx-request .htmx-indicator {
    @apply inline-block;
  }

  .htmx-request .htmx-request-hide {
    @apply hidden;
  }
}
`

const tplTailwind = `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [],
};
`

const tplPackageJSON = `{
  "private": true,
  "devDependencies": {
    "prettier": "^3.5.3",
    "tailwindcss": "^3.4.17"
  },
  "scripts": {
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "test": "npm run format:check"
  }
}
`

const tplMakefile = `.PHONY: dev build test css css-watch lint format format-check pre-commit-install ci

CAIS := $(shell command -v cais 2>/dev/null || command -v $(HOME)/go/bin/cais 2>/dev/null)

BIN := bin/server
CSS_IN := input.css
CSS_OUT := web/static/css/styles.css

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

pre-commit-install:
	pre-commit install

ci: test lint format-check

css:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server

dev: css
	$(MAKE) css-watch &
	$(CAIS) dev
`

const tplCIWorkflow = `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./... -race -count=1 -v

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v7
        with:
          version: v2.12.2

  js:
    name: JS
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: npm

      - run: npm ci

      - name: Prettier
        run: npx prettier --check .

      - name: npm test
        run: npm test
`

const tplPreCommitConfig = `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
        exclude: ^web/static/
      - id: end-of-file-fixer
        exclude: ^web/static/
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0-alpha.8
    hooks:
      - id: prettier
        exclude: ^web/static/

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        pass_filenames: false
        types: [go]

      - id: go-test
        name: go test
        entry: go test ./... -race -count=1
        language: system
        pass_filenames: false
        types: [go]

      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run ./...
        language: system
        pass_filenames: false
        types: [go]

      - id: npm-test
        name: npm test
        entry: npm test
        language: system
        pass_filenames: false
        files: \.(js|json|css|html|md|ya?ml)$
`

const tplGolangci = `version: "2"

linters:
  default: none
  enable:
    - errcheck
    - gocritic
    - govet
    - ineffassign
    - staticcheck
    - unused
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$

formatters:
  enable:
    - gofmt
    - goimports
  settings:
    goimports:
      local-prefixes:
        - {{.ModulePath}}
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
`

const tplPrettierrc = `{
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "semi": true,
  "singleQuote": false,
  "trailingComma": "es5",
  "bracketSameLine": false,
  "htmlWhitespaceSensitivity": "css",
  "overrides": [
    {
      "files": "*.html",
      "options": {
        "parser": "html"
      }
    }
  ]
}
`

const tplPrettierignore = `node_modules/
bin/
tmp/
data/
web/templates/
web/static/css/styles.css
web/static/js/htmx.min.js
package-lock.json
go.sum
`

const tplGitignore = `bin/
data/
web/static/css/styles.css
node_modules/
tmp/
.air/
*.db
.DS_Store
`
