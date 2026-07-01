package console

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/peterh/liner"
)

func (r *Repl) readLine() (string, error) {
	if r.liner != nil {
		return r.liner.Prompt(">> ")
	}

	if !r.prompted {
		_, _ = fmt.Fprint(r.out, ">> ")
		r.prompted = true
	}
	if !r.scanner.Scan() {
		_, _ = fmt.Fprintln(r.out)
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}

func (r *Repl) initInput() {
	if r.opts.In != nil {
		r.scanner = bufio.NewScanner(r.opts.in())
		return
	}
	if liner.TerminalSupported() {
		ln := liner.NewLiner()
		ln.SetCtrlCAborts(true)
		for _, line := range r.history.Lines() {
			ln.AppendHistory(line)
		}
		r.liner = ln
		return
	}
	r.scanner = bufio.NewScanner(os.Stdin)
}

func (r *Repl) closeInput() {
	if r.liner != nil {
		_ = r.liner.Close()
	}
}
