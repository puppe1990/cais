package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func nextMigrationFile(dir, slug string, dryRun bool) (relPath, num string, err error) {
	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	if !dryRun {
		if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
			return "", "", err
		}
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", "", err
	}
	maxNum := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			var n int
			if _, scanErr := fmt.Sscanf(e.Name(), "%03d_", &n); scanErr == nil && n > maxNum {
				maxNum = n
			}
		}
	}
	num = fmt.Sprintf("%03d", maxNum+1)
	relPath = filepath.Join("internal/store/migrations", num+"_"+slug+".sql")
	return relPath, num, nil
}
