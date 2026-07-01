package patch

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// InsertBeforeFuncEnd inserts Go statements before the closing brace of funcName.
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

	insertStmts, err := parseStmtList(insert)
	if err != nil {
		return nil, err
	}
	target.Body.List = append(target.Body.List, insertStmts...)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}
	return buf.Bytes(), nil
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
