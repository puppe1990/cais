package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var scaffoldOut io.Writer = os.Stdout

func setScaffoldOut(w io.Writer) {
	if w != nil {
		scaffoldOut = w
	}
}

func printfScaffold(action, rel string) {
	_, _ = fmt.Fprintf(scaffoldOut, "  %s %s\n", action, rel)
}

func writeScaffoldFile(fullPath string, content []byte, perm os.FileMode, rel string, dryRun bool) error {
	if dryRun {
		printfScaffold("create", rel)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, perm)
}

func writeScaffoldTemplate(fullPath, tpl string, data scaffoldData, rel string, dryRun bool) error {
	if dryRun {
		printfScaffold("create", rel)
		return nil
	}
	return writeTemplate(fullPath, tpl, data)
}

func updateScaffoldFile(fullPath string, content []byte, rel string, dryRun bool) error {
	if dryRun {
		printfScaffold("update", rel)
		return nil
	}
	return os.WriteFile(fullPath, content, 0o644)
}
