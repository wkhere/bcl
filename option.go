package bcl

import (
	"io"
	"os"
)

type Option func(*config)

type config struct {
	disasm bool
	trace  bool
	stats  bool
	output io.Writer
}

func makeConfig(oo []Option) (cf config) {
	cf.output = os.Stdout

	for _, o := range oo {
		o(&cf)
	}
	return cf
}

func OptDisasm(x bool) Option {
	return func(cf *config) { cf.disasm = x }
}

func OptTrace(x bool) Option {
	return func(cf *config) { cf.trace = x }
}

func OptStats(x bool) Option {
	return func(cf *config) { cf.stats = x }
}

func OptOutput(w io.Writer) Option {
	return func(cf *config) { cf.output = w }
}
