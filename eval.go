package bcl

import (
	"fmt"
	"strconv"
)

type (
	env struct {
		stage   evalStage
		varDefs map[nIdent]expr
		varMark map[nIdent]mark
		varVals map[nVarRef]any
	}

	evalStage int

	mark struct{}
)

const (
	stResolveVars evalStage = iota
	stResolveBlockFields
)

func eval(top *nTop) (bb []Block, _ error) {
	env := &env{
		varDefs: top.vars,
		varMark: make(map[nIdent]mark, len(top.vars)),
		varVals: make(map[nVarRef]any, len(top.vars)),
		stage:   stResolveVars,
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
			Kind:   string(nb.kind),
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

func resolveVar(ident nIdent, env *env) (any, error) {
	expr, ok := env.varDefs[ident]
	if !ok {
		return nil, errNoVar(ident)
	}

	env.varMark[ident] = mark{}

	v, err := expr.eval(env)
	if err != nil {
		return nil, err
	}
	env.varVals[nVarRef(ident)] = v
	return v, nil
}

func (v nVarRef) eval(env *env) (any, error) {
	switch env.stage {
	case stResolveVars:
		if val, ok := env.varVals[v]; ok {
			return val, nil
		}
		if _, mark := env.varMark[nIdent(v)]; mark {
			return nil, errRecursionLoop(v)
		}
		return resolveVar(nIdent(v), env)

	case stResolveBlockFields:
		if val, ok := env.varVals[v]; ok {
			return val, nil
		}
		return nil, errNoVar(v)

	default:
		return nil, errInvalidStage(env.stage)
	}
}

func (x nIntLit) eval(env *env) (any, error) { return int(x), nil }
func (s nStrLit) eval(env *env) (any, error) { return string(s), nil }

func (o nBinOp) eval(env *env) (any, error) {
	l, err := o.a.eval(env)
	if err != nil {
		return nil, err
	}
	r, err := o.b.eval(env)
	if err != nil {
		return nil, err
	}

	switch o.op {
	case "+":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv + rv, nil
			}

		case string:
			switch rv := r.(type) {
			case string:
				return lv + rv, nil
			case int:
				return lv + strconv.Itoa(rv), nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r}

	default:
		return nil, errUnknownOp(o.op)
	}
}

type errNoVar nIdent
type errRecursionLoop nVarRef
type errInvalidStage int
type errUnknownOp op

type errOpInvalidTypes struct {
	op   op
	x, y any
}

func (e errNoVar) Error() string {
	return fmt.Sprintf("var %s not defined", string(e))
}

func (e errRecursionLoop) Error() string {
	return fmt.Sprintf("var %s: recursion loop", string(e))
}

func (e errInvalidStage) Error() string {
	return fmt.Sprintf("invalid eval stage: %d", int(e))
}

func (e errUnknownOp) Error() string {
	return fmt.Sprintf("unknown op %v", op(e))
}

func (e *errOpInvalidTypes) Error() string {
	return fmt.Sprintf("op %q: invalid types: %T, %T", e.op, e.x, e.y)
}
