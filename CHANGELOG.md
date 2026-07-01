# Changelog

## Unreleased

### Added

- Versioned migrations (`pkg/cais/migrate`, `cais db migrate`, `cais db status`)
- CSRF protection (`middleware.CSRF`, `meta.WithCSRF`)
- Session auth in scaffold (`cais g auth`, login/logout, protected dashboard)
- `validate.Email`, SQLite production defaults, DB-aware `/health`
- `cais g auth`, smoke scaffold CI script

### Changed

- Admin auth requires `ADMIN_TOKEN` in production; Bearer header only
- Reference app aligned with `cais new` (routes.go, auth, dashboard)

### Security

- Removed admin token via query string
- Constant-time admin token comparison
