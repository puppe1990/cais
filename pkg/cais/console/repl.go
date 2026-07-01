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

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var errExit = errors.New("console exit")

type Repl struct {
	opts   Options
	interp *interp.Interpreter
	out    io.Writer
}

func New(opts Options) *Repl {
	return &Repl{opts: opts, out: opts.out()}
}

func Run(opts Options) error {
	return New(opts).Loop()
}

func (r *Repl) Loop() error {
	if err := r.initInterp(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(r.out, "=> %s console (%s)\n", r.opts.AppName, r.opts.Config.Env)
	_, _ = fmt.Fprintln(r.out, "=> Type Go expressions. Variables: store, cfg, db. Commands: help, sql, exit")
	_, _ = fmt.Fprintln(r.out, "=> Example: store.FindUserByEmail(\"demo@pulsefit.local\")")

	scanner := bufio.NewScanner(r.opts.in())
	for {
		_, _ = fmt.Fprint(r.out, ">> ")
		if !scanner.Scan() {
			_, _ = fmt.Fprintln(r.out)
			return nil
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
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
	case strings.HasPrefix(line, "sql "):
		return r.runSQL(strings.TrimSpace(strings.TrimPrefix(line, "sql")))
	default:
		return r.EvalLine(line)
	}
}

func (r *Repl) printHelp() {
	_, _ = fmt.Fprintln(r.out, "Bindings:")
	for name := range r.opts.Bindings {
		_, _ = fmt.Fprintf(r.out, "  %s\n", name)
	}
	_, _ = fmt.Fprintln(r.out, "Commands:")
	_, _ = fmt.Fprintln(r.out, "  help          show this help")
	_, _ = fmt.Fprintln(r.out, "  sql <query>   run raw SQL against db")
	_, _ = fmt.Fprintln(r.out, "  exit          leave console")
	_, _ = fmt.Fprintln(r.out, "Go examples:")
	_, _ = fmt.Fprintln(r.out, `  store.FindUserByEmail("demo@pulsefit.local")`)
	_, _ = fmt.Fprintln(r.out, `  import "fmt"; fmt.Println(cfg.DBPath)`)
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
		prelude = append(prelude, fmt.Sprintf(`%s := caisrepl.Bind(%q)`, name, name))
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
	store := r.opts.Bindings["store"]
	cfg := r.opts.Config
	db, _ := r.opts.Bindings["db"].(*sql.DB)
	extra := map[string]any{}
	for name, val := range r.opts.Bindings {
		if name == "store" || name == "cfg" || name == "db" {
			continue
		}
		extra[name] = val
	}

	return interp.Exports{
		"caisrepl/caisrepl": {
			"Store": typedProvider(store),
			"Cfg":   reflect.ValueOf(func() cais.Config { return cfg }),
			"DB":    reflect.ValueOf(func() *sql.DB { return db }),
			"Bind": reflect.ValueOf(func(name string) any {
				if v, ok := extra[name]; ok {
					return v
				}
				return nil
			}),
		},
	}
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
