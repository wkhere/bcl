package bcl

type (
	node interface {
		isnode()
	}

	expr interface {
		isexpr()
		eval(*env) (any, error)
	}

	ident string

	op string

	nIntLit int
	nStrLit string

	nVarRef ident

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

func (nIntLit) isnode() {}
func (nStrLit) isnode() {}
func (nVarRef) isnode() {}
func (nBinOp) isnode()  {}
func (nBlock) isnode()  {}
func (nTop) isnode()    {}

func (nIntLit) isexpr() {}
func (nStrLit) isexpr() {}
func (nVarRef) isexpr() {}
func (nBinOp) isexpr()  {}
