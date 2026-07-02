package cli

import "github.com/puppe1990/cais/internal/cli/patch"

// insertBeforeFunctionEnd appends statements before the closing brace of funcName.
// Uses go/ast (internal/cli/patch), not regex: generated routes often nest
// cais.IntParam and middleware groups — naive "insert before last }" corrupts the file.
func insertBeforeFunctionEnd(content, funcName, insert string) (string, error) {
	out, err := patch.InsertBeforeFuncEnd([]byte(content), funcName, insert)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
