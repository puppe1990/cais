package cli

const tplREADME = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Cais](https://github.com/puppe1990/cais): server-side HTML, HTMX, Tailwind, and SQLite.\n\n" +
	"## Stack\n\n" +
	"- Go 1.26 (net/http stdlib)\n" +
	"- html/template + HTMX 2.x\n" +
	"- Tailwind CSS 3.x\n" +
	"- SQLite (modernc.org/sqlite, no CGO)\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"cais install  # npm install + go mod tidy\n" +
	"cais dev        # http://localhost:8080\n" +
	"cais test       # full test suite\n" +
	"cais build      # bin/server\n" +
	"```\n\n" +
	"## Cais CLI\n\n" +
	"This app was scaffolded with the Cais CLI. Useful commands:\n\n" +
	"```bash\n" +
	"cais install               # npm install + go mod tidy\n" +
	"cais css                   # build Tailwind\n" +
	"cais dev                   # hot reload + tailwind watch\n" +
	"cais server                # go run ./cmd/server\n" +
	"cais console               # interactive Go REPL + SQL\n" +
	"cais g handler <name>      # handler + test + page template\n" +
	"cais g resource <name>     # model + migration + admin CRUD\n" +
	"cais g page <name>         # page template only\n" +
	"cais g migration <name>    # SQL migration file\n" +
	"cais test                  # go test ./...\n" +
	"cais doctor                # verify setup\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"GitHub Actions runs Go tests, `golangci-lint`, Prettier, and `npm test` on every push/PR to `main`.\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"Pre-commit hooks run: trailing whitespace, Prettier, `go fmt`, `go test`, `golangci-lint`, and `npm test`.\n\n" +
	"## Structure\n\n" +
	"```\n" +
	"pkg/cais/          → framework (via dependency)\n" +
	"internal/app/      → bootstrap and routes\n" +
	"internal/handlers/ → HTTP handlers\n" +
	"internal/store/    → SQLite + migrations\n" +
	"web/templates/     → HTML\n" +
	"web/static/        → CSS + JS\n" +
	"cmd/server/        → entry point\n" +
	"```\n\n" +
	"## Environment variables\n\n" +
	"| Variable  | Default         | Description      |\n" +
	"| --------- | --------------- | ---------------- |\n" +
	"| PORT      | :8080           | Server port      |\n" +
	"| DB_PATH   | ./data/app.db   | SQLite file path |\n" +
	"| ENV       | development     | Environment      |\n\n" +
	"Health check: GET /health → {\"status\":\"ok\"}\n\n" +
	"## Testing on phone (LAN)\n\n" +
	"1. Run `cais dev` and note the **LAN** URL printed at boot (e.g. `http://192.168.1.10:8080`).\n" +
	"2. Open that URL in mobile Safari/Chrome on the same Wi‑Fi.\n" +
	"3. After template or SSE changes, run `cais pwa --bump` and reinstall the PWA (or clear site data) so the service worker cache refreshes.\n" +
	"4. Run `cais doctor --mobile` to catch flash markup, font CSP, and SW cache issues.\n"

const tplREADMEBlank = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Cais](https://github.com/puppe1990/cais): server-side HTML, HTMX, Tailwind, and SQLite.\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"cais install  # npm install + go mod tidy\n" +
	"cais dev        # http://localhost:8080\n" +
	"cais test       # full test suite\n" +
	"make ci         # test + lint + format-check\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"## Add your first resource\n\n" +
	"```bash\n" +
	"cais g resource bookmark --fields title:string,url:url,notes:text?\n" +
	"```\n\n" +
	"This generates:\n" +
	"- Model, migration, admin CRUD, and public list page\n" +
	"- Tests for handlers and store\n" +
	"- Routes with admin protection\n"
