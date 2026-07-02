package cli

const tplGoMod = `module {{.ModulePath}}

go 1.26

require (
	github.com/puppe1990/cais v0.1.0
	modernc.org/sqlite v1.53.0
)
`

const tplEmptyCSS = `/* Run: cais css */\n`

const tplAir = `root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  entrypoint = ["./tmp/main"]
  delay = 1000
  exclude_dir = ["tmp", "data", "bin", "node_modules"]
  include_ext = ["go", "html"]
  stop_on_error = true

[log]
  time = false
  main_only = true

[misc]
  clean_on_exit = true
  startup_banner = ""
`

const tplWebEmbed = `package web

import "embed"

//go:embed templates/*
var Templates embed.FS
`

const tplEnvExample = `# Server
PORT=:8080
ENV=development
APP_URL=http://localhost:8080
LOCALE=en

# Database
DB_PATH=./data/app.db

# Security (required when ENV=production)
ADMIN_TOKEN=

# Reverse proxy (comma-separated IPs; trust X-Forwarded-For for client IP)
TRUSTED_PROXIES=

# SMTP (optional; password reset emails — logs to stdout when unset)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=
`
