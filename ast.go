package bcl

type (
	expr interface {
		eval(*env) (any, error)
	}

	ident string

	op string

	nIntLit  int
	nStrLit  string
	nBoolLit bool

	nVarRef ident

	nUnOp struct {
		op op
		a  expr
	}

	nBinOp struct {
		op   op
		a, b expr
	}

	nBlock struct {
		kind   ident
		name   nStrLit
		fields map[ident]expr
	}

	nTop struct {
		vars   map[ident]expr
		blocks []nBlock
	}
)
