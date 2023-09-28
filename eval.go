package bcl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

	switch o.op {
	case "-":
		switch v := x.(type) {
		case int:
			return -v, nil
		case float64:
			return -v, nil
		}
		return nil, &errOpInvalidType{o.op, x, nodeLine(o, env)}

	case "not":
		switch v := x.(type) {
		case bool:
			return !v, nil
		}
		return nil, &errOpInvalidType{o.op, x, nodeLine(o, env)}

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

	switch o.op {
	case "==":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv == rv, nil
			case float64:
				return float64(lv) == rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				return lv == rv, nil
			case int:
				return lv == float64(rv), nil
			}

		case bool:
			switch rv := r.(type) {
			case bool:
				return lv == rv, nil
			}

		case string:
			switch rv := r.(type) {
			case string:
				return lv == rv, nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

	case "!=":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv != rv, nil
			case float64:
				return float64(lv) != rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				return lv != rv, nil
			case int:
				return lv != float64(rv), nil
			}

		case bool:
			switch rv := r.(type) {
			case bool:
				return lv != rv, nil
			}

		case string:
			switch rv := r.(type) {
			case string:
				return lv != rv, nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

	case "+":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv + rv, nil
			case float64:
				return float64(lv) + rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				return lv + rv, nil
			case int:
				return lv + float64(rv), nil
			}

		case string:
			switch rv := r.(type) {
			case string:
				return lv + rv, nil
			case int:
				return lv + strconv.Itoa(rv), nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

	case "-":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv - rv, nil
			case float64:
				return float64(lv) - rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				return lv - rv, nil
			case int:
				return lv - float64(rv), nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

	case "*":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				return lv * rv, nil
			case float64:
				return float64(lv) * rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				return lv * rv, nil
			case int:
				return lv * float64(rv), nil
			}

		case string:
			switch rv := r.(type) {
			case int:
				return strings.Repeat(lv, rv), nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

	case "/":
		switch lv := l.(type) {
		case int:
			switch rv := r.(type) {
			case int:
				if rv == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lv / rv, nil
			case float64:
				if rv == 0.0 {
					return nil, fmt.Errorf("division by zero")
				}
				return float64(lv) / rv, nil
			}

		case float64:
			switch rv := r.(type) {
			case float64:
				if rv == 0.0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lv / rv, nil
			case int:
				if rv == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lv / float64(rv), nil
			}
		}

		return nil, &errOpInvalidTypes{o.op, l, r, nodeLine(o, env)}

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
