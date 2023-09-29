package bcl

import (
	"strconv"
	"strings"
)

func evalUnMinus(x any, typeErr func(any) error) (any, error) {
	switch v := x.(type) {
	case int:
		return -v, nil
	case float64:
		return -v, nil
	}
	return nil, typeErr(x)
}

func evalNot(x any, typeErr func(any) error) (any, error) {
	switch v := x.(type) {
	case bool:
		return !v, nil
	}
	return nil, typeErr(x)
}

func evalEQ(l, r any, typeErr func(any, any) error) (any, error) {
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
	return nil, typeErr(l, r)
}

func evalNE(l, r any, typeErr func(any, any) error) (any, error) {
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

	return nil, typeErr(l, r)
}

func evalPlus(l, r any, typeErr func(any, any) error) (any, error) {
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

	return nil, typeErr(l, r)
}

func evalMinus(l, r any, typeErr func(any, any) error) (any, error) {
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

	return nil, typeErr(l, r)
}

func evalMult(l, r any, typeErr func(any, any) error) (any, error) {
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

	return nil, typeErr(l, r)
}

func evalDiv(l, r any, typeErr func(any, any) error) (any, error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			if rv == 0 {
				return nil, errDivisionByZero{}
			}
			return lv / rv, nil
		case float64:
			if rv == 0.0 {
				return nil, errDivisionByZero{}
			}
			return float64(lv) / rv, nil
		}

	case float64:
		switch rv := r.(type) {
		case float64:
			if rv == 0.0 {
				return nil, errDivisionByZero{}
			}
			return lv / rv, nil
		case int:
			if rv == 0 {
				return nil, errDivisionByZero{}
			}
			return lv / float64(rv), nil
		}
	}

	return nil, typeErr(l, r)
}
