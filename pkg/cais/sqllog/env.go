package sqllog

// EnabledForEnv reports whether SQL query logging should run for the given app env.
func EnabledForEnv(env string) bool {
	return env == "development"
}

// ConfigForEnv returns development-friendly defaults (enabled + JSON lines for agents).
// JSON avoids regex on Rails-style SQL text when tooling parses query duration and args.
func ConfigForEnv(env string) Config {
	return Config{
		Enabled: EnabledForEnv(env),
		JSON:    EnabledForEnv(env),
	}
}
