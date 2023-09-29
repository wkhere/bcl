package bcl

import (
	"errors"
	"fmt"
)

type (
	env struct {
		stage   evalStage
		varDefs map[ident]expr
		varMark map[ident]mark
		varVals map[ident]any
		lc      lineCalc
	}

	evalStage int

	mark struct{}
)

const (
	stResolveVars evalStage = iota
	stResolveBlockFields
)

func eval(top *nTop, lc lineCalc) (bb []Block, _ error) {
	env := &env{
		varDefs: top.vars,
		varMark: make(map[ident]mark, len(top.vars)),
		varVals: make(map[ident]any, len(top.vars)),
		stage:   stResolveVars,
		lc:      lc,
	}

	for ident, _ := range env.varDefs {
		_, err := resolveVar(ident, env)
		if err != nil {
			return nil, err
		}
	}

	env.stage = stResolveBlockFields
	bb = make([]Block, 0, len(top.blocks))

	for _, nb := range top.blocks {
		b := Block{
			Type:   string(nb.typ),
			Name:   string(nb.name),
			Fields: make(map[string]any, len(nb.fields)),
		}
		for ident, expr := range nb.fields {
			v, err := expr.eval(env)
			if err != nil {
				return bb, err
			}
			b.Fields[string(ident)] = v
		}
		bb = append(bb, b)
	}

	return bb, nil
}

func resolveVar(ident ident, env *env) (any, error) {
	expr, ok := env.varDefs[ident]
	if !ok {
		return nil, errNoVar{ident, 0}
	}

	env.varMark[ident] = mark{}

	v, err := expr.eval(env)
	if err != nil {
		return nil, err
	}
	env.varVals[ident] = v
	return v, nil
}

func (v nVarRef) eval(env *env) (any, error) {
	switch env.stage {
	case stResolveVars:
		if val, ok := env.varVals[v.ident]; ok {
			return val, nil
		}
		if _, mark := env.varMark[v.ident]; mark {
			return nil, errCycle{v.ident, nodeLine(v, env)}
		}
		val, err := resolveVar(v.ident, env)
		var noVar errNoVar
		if errors.As(err, &noVar) {
			noVar.line = nodeLine(v, env)
			err = noVar
		}
		return val, err

	case stResolveBlockFields:
		if val, ok := env.varVals[v.ident]; ok {
			return val, nil
		}
		return nil, errNoVar{v.ident, nodeLine(v, env)}

	default:
		return nil, errInvalidStage(env.stage)
	}
}

func (x nIntLit) eval(env *env) (any, error)   { return x.int, nil }
func (x nFloatLit) eval(env *env) (any, error) { return x.float64, nil }
func (s nStrLit) eval(env *env) (any, error)   { return s.string, nil }
func (b nBoolLit) eval(env *env) (any, error)  { return b.bool, nil }

func (o nUnOp) eval(env *env) (any, error) {
	x, err := o.a.eval(env)
	if err != nil {
		return nil, err
	}

	t := func(x any) error {
		return &errOpInvalidType{o.op, x, nodeLine(o, env)}
	}

	switch o.op {
	case "-":
		return evalUnMinus(x, t)

	case "not":
		return evalNot(x, t)

	default:
		return nil, errUnknownOp{"unary " + o.op, nodeLine(o, env)}
	}
}

func (o nBinOp) eval(env *env) (any, error) {
	l, err := o.a.eval(env)
	if err != nil {
		return nil, err
	}
	r, err := o.b.eval(env)
	if err != nil {
		return nil, err
	}

	t := func(l, r any) error {
		return &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}
	}

	switch o.op {
	case "==":
		return evalEQ(l, r, t)

	case "!=":
		return evalNE(l, r, t)

	case "<":
		return evalLT(l, r, t)

	case "<=":
		return evalLE(l, r, t)

	case ">":
		return evalGT(l, r, t)

	case ">=":
		return evalGE(l, r, t)

	case "+":
		return evalPlus(l, r, t)

	case "-":
		return evalMinus(l, r, t)

	case "*":
		return evalMult(l, r, t)

	case "/":
		res, err := evalDiv(l, r, t)
		if errdiv, ok := err.(errDivisionByZero); ok {
			errdiv.line = nodeLine(o, env)
			return res, errdiv
		}
		return res, err

	default:
		return nil, errUnknownOp{"binary " + o.op, nodeLine(o, env)}
	}
}

func nodeLine(n node, env *env) int {
	return env.lc.lineAt(n.getpos())
}

type (
	errInvalidStage int

	errNoVar struct {
		ident
		line int
	}
	errCycle struct {
		ident
		line int
	}
	errUnknownOp struct {
		op
		line int
	}
	errOpInvalidType struct {
		op   op
		x    any
		line int
	}
	errOpInvalidTypes struct {
		op   op
		x, y any
		line int
	}

	errDivisionByZero struct{ line int }
)

func (e errInvalidStage) Error() string {
	return fmt.Sprintf("invalid eval stage: %d", int(e))
}

func (e errNoVar) Error() string {
	return fmt.Sprintf("line %d: var %s not defined", e.line, e.ident)
}

func (e errCycle) Error() string {
	return fmt.Sprintf("line %d: var %s: cycle detected", e.line, e.ident)
}

func (e errUnknownOp) Error() string {
	return fmt.Sprintf("line %d: unknown op %q", e.line, e.op)
}

func (e *errOpInvalidType) Error() string {
	return fmt.Sprintf("line %d: op %q: invalid type: %T", e.line, e.op, e.x)
}

func (e *errOpInvalidTypes) Error() string {
	return fmt.Sprintf(
		"line %d: op %q: invalid types: %T, %T",
		e.line, e.op, e.x, e.y,
	)
}

func (e errDivisionByZero) Error() string {
	return fmt.Sprintf("line %d: division by zero", e.line)
}
