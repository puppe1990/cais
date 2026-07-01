package patch

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// InsertBeforeFuncEnd inserts Go statements before the closing brace of funcName.
// The source file is not reformatted so qualified calls such as cais.IntParam stay intact.
func InsertBeforeFuncEnd(src []byte, funcName, insert string) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "routes.go", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var target *ast.FuncDecl
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != funcName {
			continue
		}
		target = fn
		break
	}
	if target == nil || target.Body == nil {
		return nil, fmt.Errorf("function %q not found", funcName)
	}

	if _, err := parseStmtList(insert); err != nil {
		return nil, err
	}

	// Body.End() is the position after the closing brace; insert immediately before '}'.
	closeBrace := target.Body.End() - 1
	pos := fset.Position(closeBrace).Offset
	if pos <= 0 || pos > len(src) || src[pos] != '}' {
		return nil, fmt.Errorf("invalid insert position for %q", funcName)
	}

	out := make([]byte, 0, len(src)+len(insert))
	out = append(out, src[:pos]...)
	out = append(out, insert...)
	out = append(out, src[pos:]...)
	return out, nil
}

func parseStmtList(insert string) ([]ast.Stmt, error) {
	wrapped := "package p\nfunc _() {\n" + insert + "\n}"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "insert.go", wrapped, 0)
	if err != nil {
		return nil, fmt.Errorf("parse insert: %w", err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Body != nil {
			return fn.Body.List, nil
		}
	}
	return nil, fmt.Errorf("parse insert: no statements")
}
