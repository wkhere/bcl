package bcl

type Block struct {
	Kind, Name string
	Fields     map[string]any
}

func Interp(input []byte) ([]Block, error) {
	top, err := parse(input)
	if err != nil {
		return nil, err

	}
	return eval(&top)
}
