package bcl

type Option func(*config)

type config struct {
	disasm bool
	trace  bool
}

func makeConfig(oo []Option) (cf config) {
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
