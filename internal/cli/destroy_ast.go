package cli

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// #168: destroy unpatches used line scanning — dropping any line containing
// /admin/<plural> ate user comments and custom routes, prefix matching removed
// lookalike methods, and brace counting broke on braces inside string literals.
// These helpers cut removals by go/ast node offsets so only exact generated
// nodes are removed. Byte-range cutting (not format.Node) keeps user formatting.

type span struct{ start, end int }

func cutSpans(src string, spans []span) string {
	if len(spans) == 0 {
		return src
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	var b strings.Builder
	prev := 0
	for _, s := range spans {
		b.WriteString(src[prev:s.start])
		prev = s.end
	}
	b.WriteString(src[prev:])
	return b.String()
}

// spanOf converts an ast.Node to source byte offsets covering the whole line:
// leading indentation is pulled in so removals don't leave orphan tabs, and the
// trailing newline is consumed so cuts don't accumulate blank lines.
func spanOf(fset *token.FileSet, n ast.Node, src string) span {
	s := span{fset.Position(n.Pos()).Offset, fset.Position(n.End()).Offset}
	for s.start > 0 && (src[s.start-1] == '\t' || src[s.start-1] == ' ') {
		s.start--
	}
	if s.end < len(src) && src[s.end] == '\n' {
		s.end++
	}
	return s
}

func parseGo(src, filename string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	return fset, file, nil
}

// removeStmtsFromFunc cuts top-level statements of the named function for which
// drop returns true.
func removeStmtsFromFunc(src, funcDeclName string, drop func(ast.Stmt) bool) (string, error) {
	fset, file, err := parseGo(src, "routes.go")
	if err != nil {
		return "", err
	}
	var spans []span
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != funcDeclName || fd.Body == nil {
			continue
		}
		for _, st := range fd.Body.List {
			if drop(st) {
				spans = append(spans, spanOf(fset, st, src))
			}
		}
	}
	return cutSpans(src, spans), nil
}

// removeDeclsByName cuts every top-level FuncDecl whose name is exactly in names.
func removeDeclsByName(src string, names map[string]bool) (string, error) {
	fset, file, err := parseGo(src, "store.go")
	if err != nil {
		return "", err
	}
	var spans []span
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if ok && names[fd.Name.Name] {
			spans = append(spans, spanOf(fset, decl, src))
		}
	}
	return cutSpans(src, spans), nil
}

// removeInterfaceMethods cuts fields with exactly matching names from the named
// interface; other interfaces sharing method names are untouched.
func removeInterfaceMethods(src, ifaceName string, names map[string]bool) (string, error) {
	fset, file, err := parseGo(src, "store.go")
	if err != nil {
		return "", err
	}
	var spans []span
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != ifaceName {
				continue
			}
			iface, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, field := range iface.Methods.List {
				if len(field.Names) == 0 || !names[field.Names[0].Name] {
					continue
				}
				spans = append(spans, spanOf(fset, field, src))
			}
		}
	}
	return cutSpans(src, spans), nil
}

// litStrings returns the decoded-ish STRING literal values under n (raw quoted).
func litStrings(n ast.Node) []string {
	var out []string
	ast.Inspect(n, func(node ast.Node) bool {
		if lit, ok := node.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			out = append(out, strings.Trim(lit.Value, "`\""))
		}
		return true
	})
	return out
}

// isGeneratedRouteStmt reports whether stmt is one of the resource generator's
// route statements: the handler-var assignments, the public r.Get/r.Post calls,
// or the whole admin r.Group block (matched via its exact path literals).
func isGeneratedRouteStmt(st ast.Stmt, pubVar, adminVar string, pubPaths, adminPaths map[string]bool) bool {
	switch s := st.(type) {
	case *ast.AssignStmt:
		for _, lhs := range s.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && (id.Name == pubVar || id.Name == adminVar) {
				return true
			}
		}
	case *ast.ExprStmt:
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		recv, ok := sel.X.(*ast.Ident)
		if !ok || recv.Name != "r" {
			return false
		}
		switch sel.Sel.Name {
		case "Get", "Post", "Delete":
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					return pubPaths[strings.Trim(lit.Value, "`\"")]
				}
			}
		case "Group":
			for _, v := range litStrings(call) {
				if adminPaths[v] {
					return true
				}
			}
		}
	}
	return false
}
