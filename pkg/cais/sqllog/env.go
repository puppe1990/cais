package sqllog

// EnabledForEnv reports whether SQL query logging should run for the given app env.
func EnabledForEnv(env string) bool {
	return env == "development"
}
