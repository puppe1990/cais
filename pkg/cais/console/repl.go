package console

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/peterh/liner"
	"github.com/puppe1990/cais/pkg/cais"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var errExit = errors.New("console exit")

type Repl struct {
	opts     Options
	interp   *interp.Interpreter
	out      io.Writer
	history  *History
	scanner  *bufio.Scanner
	liner    *liner.State
	prompted bool
}

func New(opts Options) *Repl {
	return &Repl{opts: opts, out: opts.out(), history: NewHistory()}
}

func Run(opts Options) error {
	return New(opts).Loop()
}

func (r *Repl) Loop() error {
	r.initInput()
	defer r.closeInput()

	if err := r.initInterp(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(r.out, "=> %s console (%s)\n", r.opts.AppName, r.opts.Config.Env)
	_, _ = fmt.Fprintln(r.out, "=> Variables: store, cfg, db + custom bindings. Commands: help, sql, reload, history, exit")
	_, _ = fmt.Fprintln(r.out, "=> Example: store.FindUserByEmail(\"demo@pulsefit.local\")")

	for {
		line, err := r.readLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if r.shouldRecord(line) {
			r.recordHistory(line)
		}
		if err := r.HandleLine(line); err != nil {
			if errors.Is(err, errExit) {
				return nil
			}
			_, _ = fmt.Fprintf(r.out, "Error: %v\n", err)
		}
	}
}

func (r *Repl) HandleLine(line string) error {
	switch {
	case line == "exit", line == "quit":
		_, _ = fmt.Fprintln(r.out, "Bye!")
		return errExit
	case line == "help":
		r.printHelp()
		return nil
	case line == "history":
		r.printHistory()
		return nil
	case line == "reload":
		return r.reload()
	case strings.HasPrefix(line, "!"):
		return r.recallHistory(line)
	case strings.HasPrefix(line, "sql "):
		return r.runSQL(strings.TrimSpace(strings.TrimPrefix(line, "sql")))
	default:
		return r.EvalLine(line)
	}
}

func (r *Repl) shouldRecord(line string) bool {
	switch {
	case line == "", line == "help", line == "history", line == "exit", line == "quit", line == "reload":
		return false
	case strings.HasPrefix(line, "!"):
		return false
	default:
		return true
	}
}

func (r *Repl) recordHistory(line string) {
	r.history.Add(line)
	if r.liner != nil {
		r.liner.AppendHistory(line)
	}
}

func (r *Repl) printHistory() {
	lines := r.history.List()
	if len(lines) == 0 {
		_, _ = fmt.Fprintln(r.out, "(empty)")
		return
	}
	for i, line := range lines {
		_, _ = fmt.Fprintf(r.out, "%4d  %s\n", i+1, line)
	}
}

func (r *Repl) recallHistory(line string) error {
	var cmd string
	switch line {
	case "!!":
		cmd = r.history.Last()
		if cmd == "" {
			return fmt.Errorf("no previous command")
		}
	default:
		var n int
		if _, err := fmt.Sscanf(line, "!%d", &n); err != nil || n < 1 {
			return fmt.Errorf("usage: !N for numbered recall or !! for last command")
		}
		cmd = r.history.At(n)
		if cmd == "" {
			return fmt.Errorf("history entry %d not found", n)
		}
	}
	_, _ = fmt.Fprintf(r.out, ">> %s\n", cmd)
	return r.HandleLine(cmd)
}

func (r *Repl) printHelp() {
	_, _ = fmt.Fprintln(r.out, "Bindings:")
	names := bindingNames(r.opts.Bindings)
	for _, name := range names {
		_, _ = fmt.Fprintf(r.out, "  %s\n", name)
	}
	_, _ = fmt.Fprintln(r.out, "  cfg           (from cais.Load)")
	_, _ = fmt.Fprintln(r.out, "Commands:")
	_, _ = fmt.Fprintln(r.out, "  help          show this help")
	_, _ = fmt.Fprintln(r.out, "  sql <query>   run raw SQL against db")
	_, _ = fmt.Fprintln(r.out, "  reload        refresh bindings (reconnect store/db)")
	_, _ = fmt.Fprintln(r.out, "  history       list command history")
	_, _ = fmt.Fprintln(r.out, "  !N / !!       rerun history entry")
	_, _ = fmt.Fprintln(r.out, "  exit          leave console")
	_, _ = fmt.Fprintln(r.out, "Go examples:")
	_, _ = fmt.Fprintln(r.out, `  store.FindUserByEmail("demo@pulsefit.local")`)
	_, _ = fmt.Fprintln(r.out, `  import "fmt"; fmt.Println(cfg.DBPath)`)
}

func bindingNames(bindings map[string]any) []string {
	seen := map[string]bool{"cfg": true}
	var names []string
	for _, key := range []string{"store", "db"} {
		if _, ok := bindings[key]; ok {
			names = append(names, key)
			seen[key] = true
		}
	}
	for name := range bindings {
		if seen[name] {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (r *Repl) reload() error {
	if r.opts.Reload == nil {
		return fmt.Errorf("reload not configured — add Reload to console.Options")
	}
	bindings, err := r.opts.Reload()
	if err != nil {
		return err
	}
	r.opts.Bindings = bindings
	if err := r.initInterp(); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(r.out, "Reloaded!")
	return nil
}

func (r *Repl) runSQL(query string) error {
	raw, ok := r.opts.Bindings["db"].(*sql.DB)
	if !ok || raw == nil {
		return fmt.Errorf("db binding not available")
	}
	rows, err := raw.Query(query)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	out, err := formatSQLRows(rows)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(r.out, out)
	return nil
}

func (r *Repl) initInterp() error {
	i := interp.New(interp.Options{Stdout: r.out, Stderr: r.out})
	if err := i.Use(stdlib.Symbols); err != nil {
		return err
	}
	if err := i.Use(r.bindingSymbols()); err != nil {
		return err
	}

	prelude := []string{
		`import "caisrepl/caisrepl"`,
		`store := caisrepl.Store()`,
		`cfg := caisrepl.Cfg()`,
		`db := caisrepl.DB()`,
	}
	for name := range r.opts.Bindings {
		if name == "store" || name == "cfg" || name == "db" {
			continue
		}
		prelude = append(prelude, fmt.Sprintf(`%s := caisrepl.%s()`, name, bindingExportName(name)))
	}
	for _, stmt := range prelude {
		if _, err := i.Eval(stmt); err != nil {
			return fmt.Errorf("console init: %w", err)
		}
	}
	r.interp = i
	return nil
}

func (r *Repl) bindingSymbols() interp.Exports {
	bindings := map[string]any{}
	for name, val := range r.opts.Bindings {
		bindings[name] = val
	}
	cfg := r.opts.Config

	syms := map[string]reflect.Value{
		"Store": typedProvider(bindings["store"]),
		"Cfg":   reflect.ValueOf(func() cais.Config { return cfg }),
		"DB": reflect.ValueOf(func() *sql.DB {
			db, _ := bindings["db"].(*sql.DB)
			return db
		}),
	}
	for name, val := range bindings {
		if name == "store" || name == "cfg" || name == "db" {
			continue
		}
		syms[bindingExportName(name)] = typedProvider(val)
	}

	return interp.Exports{"caisrepl/caisrepl": syms}
}

func bindingExportName(name string) string {
	parts := strings.Split(name, "_")
	var b strings.Builder
	b.WriteString("Get")
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}

func typedProvider(v any) reflect.Value {
	if v == nil {
		return reflect.ValueOf(func() any { return nil })
	}
	t := reflect.TypeOf(v)
	fn := reflect.FuncOf(nil, []reflect.Type{t}, false)
	return reflect.MakeFunc(fn, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(v)}
	})
}

func (r *Repl) EvalLine(line string) error {
	if r.interp == nil {
		if err := r.initInterp(); err != nil {
			return err
		}
	}

	v, err := r.interp.Eval(line)
	if err != nil {
		if _, err = r.interp.Eval(line + "\n"); err != nil {
			return err
		}
		return nil
	}
	if v.IsValid() && v.CanInterface() {
		r.printValue(v.Interface())
	}
	return nil
}

func (r *Repl) printValue(v any) {
	switch val := v.(type) {
	case nil:
		_, _ = fmt.Fprintln(r.out, "nil")
	case string:
		_, _ = fmt.Fprintf(r.out, "%q\n", val)
	case error:
		if val == nil {
			_, _ = fmt.Fprintln(r.out, "nil")
			return
		}
		_, _ = fmt.Fprintf(r.out, "%v\n", val)
	default:
		if b, err := json.MarshalIndent(val, "", "  "); err == nil {
			_, _ = fmt.Fprintln(r.out, string(b))
			return
		}
		_, _ = fmt.Fprintf(r.out, "%#v\n", val)
	}
}
