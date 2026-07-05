package cli

const tplInputCSS = `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    @apply font-sans antialiased text-slate-900 bg-slate-50;
  }
}

@layer utilities {
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }

  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
}

@layer components {
  .cais-nav-icon {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
  }

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

  .cais-toast-enter {
    animation: cais-toast-in 200ms ease-out;
  }

  @keyframes cais-toast-in {
    from {
      opacity: 0;
      transform: translate(-50%, -0.75rem);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0);
    }
  }

  .htmx-request .cais-skeleton {
    @apply animate-pulse bg-slate-200 rounded-lg;
  }

  .cais-auth-screen {
    @apply min-h-screen bg-gradient-to-br from-indigo-50 via-white to-violet-100;
  }

  .cais-password-wrap {
    @apply relative;
  }

  .cais-password-wrap input {
    padding-right: 2.5rem;
  }

  .cais-password-toggle {
    @apply absolute right-0 top-0 flex h-full items-center px-3 text-slate-400 hover:text-slate-600;
    border: none;
    background: transparent;
    cursor: pointer;
  }

  .cais-password-toggle svg {
    width: 1rem;
    height: 1rem;
  }

  .relative > [data-cais-password-toggle] {
    @apply absolute right-0 top-0 flex h-full items-center px-3 text-slate-400 hover:text-slate-600;
    border: none;
    background: transparent;
    cursor: pointer;
  }

  .relative > input[type="password"] {
    padding-right: 2.5rem;
  }

  .cais-chat-messages-wrap {
    position: relative;
    min-height: 0;
  }

  #chat-messages {
    overflow-anchor: none;
    -webkit-overflow-scrolling: touch;
  }

  .cais-chat-scroll-down {
    position: absolute;
    bottom: 0.75rem;
    left: 50%;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 9999px;
    border: 1px solid rgb(226 232 240);
    background: rgb(255 255 255);
    color: rgb(79 70 229);
    box-shadow:
      0 10px 15px -3px rgb(0 0 0 / 0.1),
      0 4px 6px -4px rgb(0 0 0 / 0.1);
    transform: translateX(-50%) translateY(0.5rem);
    opacity: 0;
    pointer-events: none;
    transition:
      opacity 0.2s ease,
      transform 0.2s ease;
  }

  .cais-chat-scroll-down:not(.hidden) {
    opacity: 1;
    pointer-events: auto;
    transform: translateX(-50%) translateY(0);
  }

  .cais-chat-scroll-down:active {
    transform: translateX(-50%) scale(0.96);
  }

  .cais-chat-bubble {
    overflow-wrap: anywhere;
    word-break: break-word;
    white-space: pre-wrap;
  }

  .cais-msg-time {
    font-size: 0.625rem;
    line-height: 1rem;
    font-weight: 600;
    letter-spacing: 0.01em;
    color: rgb(148 163 184);
  }

  .cais-msg-user .cais-msg-time {
    color: rgb(129 140 248);
  }

  .cais-thinking-dots {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    width: 1.5rem;
  }

  .cais-thinking-dots span {
    display: block;
    width: 0.35rem;
    height: 0.35rem;
    border-radius: 9999px;
    background: rgb(148 163 184);
    animation: cais-thinking-bounce 1.2s ease-in-out infinite;
  }

  .cais-thinking-dots span:nth-child(2) {
    animation-delay: 0.15s;
  }

  .cais-thinking-dots span:nth-child(3) {
    animation-delay: 0.3s;
  }

  @keyframes cais-thinking-bounce {
    0%,
    80%,
    100% {
      transform: translateY(0);
      opacity: 0.4;
    }
    40% {
      transform: translateY(-0.2rem);
      opacity: 1;
    }
  }
}
`

const tplTailwind = `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  safelist: [
    "cais-password-wrap",
    "cais-password-toggle",
    "cais-chat-scroll-down",
    "cais-thinking",
    "cais-thinking-dots",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["ui-sans-serif", "system-ui", "-apple-system", "Segoe UI", "sans-serif"],
        display: ["ui-sans-serif", "system-ui", "-apple-system", "Segoe UI", "sans-serif"],
        mono: ['"JetBrains Mono"', "ui-monospace", "SFMono-Regular", "monospace"],
      },
      boxShadow: {
        "2xs": "0 1px 2px 0 rgb(0 0 0 / 0.05)",
        xs: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
      },
    },
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
web/static/js/idiomorph-ext.min.js
web/static/js/sse-ext.min.js
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
