package sqllog

import "testing"

func TestConfigForEnv_development(t *testing.T) {
	cfg := ConfigForEnv("development")
	if !cfg.Enabled || !cfg.JSON {
		t.Fatalf("cfg = %+v, want enabled JSON logging", cfg)
	}
}

func TestConfigForEnv_production(t *testing.T) {
	cfg := ConfigForEnv("production")
	if cfg.Enabled || cfg.JSON {
		t.Fatalf("cfg = %+v, want disabled", cfg)
	}
}
