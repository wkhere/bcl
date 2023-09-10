package bcl

import (
	"fmt"
	"strconv"
)

type (
	env struct {
		stage   evalStage
		varDefs map[ident]expr
		varMark map[ident]mark
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
		varMark: make(map[ident]mark, len(top.vars)),
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

func resolveVar(ident ident, env *env) (any, error) {
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
		if _, mark := env.varMark[ident(v)]; mark {
			return nil, errCycle(v)
		}
		return resolveVar(ident(v), env)

	case stResolveBlockFields:
		if val, ok := env.varVals[v]; ok {
			return val, nil
		}
		return nil, errNoVar(v)

	default:
		return nil, errInvalidStage(env.stage)
	}
}

func (x nIntLit) eval(env *env) (any, error)  { return int(x), nil }
func (s nStrLit) eval(env *env) (any, error)  { return string(s), nil }
func (b nBoolLit) eval(env *env) (any, error) { return bool(b), nil }

func (o nUnOp) eval(env *env) (any, error) {
	x, err := o.a.eval(env)
	if err != nil {
		return nil, err
	}

	switch o.op {
	case "-":
		switch v := x.(type) {
		case int:
			return -v, nil
		}
		return nil, &errOpInvalidType{o.op, x}

	case "not":
		switch v := x.(type) {
		case bool:
			return !v, nil
		}
		return nil, &errOpInvalidType{o.op, x}

	default:
		return nil, errUnknownOp("unary " + o.op)
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

	case "-":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv - rv, nil
			}
		}
		return nil, &errOpInvalidTypes{o.op, l, r}

	case "*":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv * rv, nil
			}
		}
		return nil, &errOpInvalidTypes{o.op, l, r}

	case "/":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv / rv, nil
			}
		}
		return nil, &errOpInvalidTypes{o.op, l, r}

	default:
		return nil, errUnknownOp("binary " + o.op)
	}
}

type errNoVar ident
type errCycle nVarRef
type errInvalidStage int
type errUnknownOp string

type errOpInvalidType struct {
	op op
	x  any
}
type errOpInvalidTypes struct {
	op   op
	x, y any
}

func (e errNoVar) Error() string {
	return fmt.Sprintf("var %s not defined", string(e))
}

func (e errCycle) Error() string {
	return fmt.Sprintf("var %s: cycle detected", string(e))
}

func (e errInvalidStage) Error() string {
	return fmt.Sprintf("invalid eval stage: %d", int(e))
}

func (e errUnknownOp) Error() string {
	return fmt.Sprintf("unknown op %q", string(e))
}

func (e *errOpInvalidType) Error() string {
	return fmt.Sprintf("op %q: invalid type: %T", e.op, e.x)
}

func (e *errOpInvalidTypes) Error() string {
	return fmt.Sprintf("op %q: invalid types: %T, %T", e.op, e.x, e.y)
}
