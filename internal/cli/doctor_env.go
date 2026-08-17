package cli

import (
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/dotenv"
)

func isProduction(dir string) bool {
	return resolveEnvVar(dir, "ENV") == "production"
}

func resolveEnvVar(dir, key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		return ""
	}
	return dotenv.Parse(data)[key]
}

func checkAdminToken(dir string) doctorCheck {
	if resolveEnvVar(dir, "ADMIN_TOKEN") != "" {
		return doctorCheck{Name: "ADMIN_TOKEN", OK: true}
	}
	return doctorCheck{
		Name:     "ADMIN_TOKEN",
		Optional: true,
		Detail:   "required when ENV=production",
		FixHint:  "set ADMIN_TOKEN in .env",
	}
}

func checkAppURL(dir string) doctorCheck {
	if resolveEnvVar(dir, "APP_URL") != "" {
		return doctorCheck{Name: "APP_URL", OK: true}
	}
	return doctorCheck{
		Name:     "APP_URL",
		Optional: true,
		Detail:   "required when ENV=production",
		FixHint:  "set APP_URL in .env",
	}
}

func hasAuthHandler(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go"))
	return err == nil
}

func checkSMTP(dir string) doctorCheck {
	if resolveEnvVar(dir, "SMTP_HOST") != "" && resolveEnvVar(dir, "SMTP_FROM") != "" {
		return doctorCheck{Name: "SMTP", OK: true}
	}
	return doctorCheck{
		Name:     "SMTP",
		Optional: true,
		Detail:   "password reset emails log to stdout without SMTP_HOST/SMTP_FROM",
		FixHint:  "set SMTP_HOST and SMTP_FROM in .env for outbound mail",
	}
}
