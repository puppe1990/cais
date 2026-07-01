package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var seedCallPattern = regexp.MustCompile(`s\.(SeedDemo\w+|Insert\w+)\(`)

func listSeedsInDir(dir string) ([]string, error) {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	matches := seedCallPattern.FindAllStringSubmatch(string(body), -1)
	seen := make(map[string]struct{})
	var items []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		if _, ok := seen[m[1]]; ok {
			continue
		}
		seen[m[1]] = struct{}{}
		items = append(items, m[1])
	}
	sort.Strings(items)
	return items, nil
}
