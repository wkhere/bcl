package bcl

import "fmt"

type value any

func isInt(v value) bool {
	_, ok := v.(int)
	return ok
}

func isFloat(v value) bool {
	_, ok := v.(float64)
	return ok
}

func isNumber(v value) bool {
	return isInt(v) || isFloat(v)
}

func isString(v value) bool {
	_, ok := v.(string)
	return ok
}

func isBool(v value) bool {
	_, ok := v.(bool)
	return ok
}

func isFalsey(v value) bool {
	switch x := v.(type) {
	case bool:
		return !x
	case string:
		return x == ""
	default:
		return x == nil
	}
}

func isTruthy(v value) bool { return !isFalsey(v) }

func vtype(v value) string {
	switch v.(type) {
	case int:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("unknown:%T", v)
	}
}
