package bcl

type (
	node interface {
		getpos() int
	}

	expr interface {
		node
		eval(*env) (any, error)
	}

	ident string
	op    string
	pos   int

	nIntLit struct {
		int
		pos
	}
	nFloatLit struct {
		float64
		pos
	}
	nStrLit struct {
		string
		pos
	}
	nBoolLit struct {
		bool
		pos
	}

	nVarRef struct {
		ident
		pos
	}

	nUnOp struct {
		op op
		a  expr
	}

	nBinOp struct {
		op   op
		a, b expr
	}

	nSCOp struct { // SC = Short Circuit
		op   op
		a, b expr
	}

	nBlock struct {
		typ    ident
		name   ident
		fields map[ident]expr
		pos
	}

	nTop struct {
		vars   map[ident]expr
		blocks []nBlock
		pos
	}
)

func (p pos) getpos() int    { return int(p) }
func (o nUnOp) getpos() int  { return o.a.getpos() }
func (o nSCOp) getpos() int  { return o.a.getpos() }
func (o nBinOp) getpos() int { return o.a.getpos() }
