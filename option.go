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
	logw   io.Writer
}

func makeConfig(oo []Option) (cf config) {
	cf = config{
		output: os.Stdout,
		logw:   os.Stderr,
	}

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

func OptLogger(w io.Writer) Option {
	return func(cf *config) { cf.logw = w }
}
