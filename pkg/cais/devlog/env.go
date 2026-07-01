package devlog

// Enabled reports whether the /logs viewer should be available.
func Enabled(env string) bool {
	return env == "development"
}
