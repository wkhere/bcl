package main

type (
	node interface {
		isnode()
	}

	expr interface {
		isexpr()
	}

	nEmpty struct{}

	nIntLit int
	nStrLit string

	nIdent  string
	nVarRef nIdent

	nBinOp struct {
		op   rune
		a, b expr
	}

	nTunnel struct {
		name   nStrLit
		fields map[nIdent]expr
	}

	nTop struct {
		vars    map[nIdent]expr
		tunnels []nTunnel
	}
)

func (nEmpty) isnode()  {}
func (nIntLit) isnode() {}
func (nStrLit) isnode() {}
func (nIdent) isnode()  {}
func (nVarRef) isnode() {}
func (nBinOp) isnode()  {}
func (nTunnel) isnode() {}
func (nTop) isnode()    {}

func (nIntLit) isexpr() {}
func (nStrLit) isexpr() {}
func (nVarRef) isexpr() {}
func (nBinOp) isexpr()  {}
