package bcl

type (
	node interface {
		isnode()
	}

	expr interface {
		isexpr()
		eval(*env) (any, error)
	}

	op string

	nIntLit int
	nStrLit string

	nIdent  string
	nVarRef nIdent

	nBinOp struct {
		op   op
		a, b expr
	}

	nBlock struct {
		kind   nIdent
		name   nStrLit
		fields map[nIdent]expr
	}

	nTop struct {
		vars   map[nIdent]expr
		blocks []nBlock
	}
)

func (nIntLit) isnode() {}
func (nStrLit) isnode() {}
func (nIdent) isnode()  {}
func (nVarRef) isnode() {}
func (nBinOp) isnode()  {}
func (nBlock) isnode()  {}
func (nTop) isnode()    {}

func (nIntLit) isexpr() {}
func (nStrLit) isexpr() {}
func (nVarRef) isexpr() {}
func (nBinOp) isexpr()  {}
