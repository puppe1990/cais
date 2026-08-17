package dotenv

import (
	"fmt"
	"os"
)

// LoadFile sets KEY=value pairs from path into the process env when the key
// is not already set. Missing file is a no-op so tests, systemd, and CI win.
func LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("dotenv: %w", err)
	}
	for key, val := range Parse(data) {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("dotenv set %s: %w", key, err)
		}
	}
	return nil
}
