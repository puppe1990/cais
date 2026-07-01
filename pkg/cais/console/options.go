package console

import (
	"io"
	"os"

	"github.com/puppe1990/cais/pkg/cais"
)

type Options struct {
	AppName  string
	Config   cais.Config
	Bindings map[string]any
	In       io.Reader
	Out      io.Writer
}

func (o Options) out() io.Writer {
	if o.Out != nil {
		return o.Out
	}
	return os.Stdout
}

func (o Options) in() io.Reader {
	if o.In != nil {
		return o.In
	}
	return os.Stdin
}
