package cli

import "github.com/puppe1990/cais/internal/cli/patch"

// insertBeforeFunctionEnd appends statements to funcName using go/ast so nested
// blocks and cais.IntParam calls survive gofmt.
func insertBeforeFunctionEnd(content, funcName, insert string) (string, error) {
	out, err := patch.InsertBeforeFuncEnd([]byte(content), funcName, insert)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
