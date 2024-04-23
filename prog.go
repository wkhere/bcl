package bcl

type prog struct {
	name      string
	code      []byte
	positions []int
	constants []value
	lineFmt   func(pos int) string
}

func newProg(name string) *prog {
	const (
		codeCap      = 64
		constantsCap = 8
	)
	return &prog{
		name:      name,
		code:      make([]byte, 0, codeCap),
		positions: make([]int, 0, codeCap),
		constants: make([]value, 0, constantsCap),
	}
}

func (p *prog) write(b byte, pos int) {
	p.code = append(p.code, b)
	p.positions = append(p.positions, pos)
}

func (p *prog) addConst(v value) (int, error) {
	if len(p.constants) == 255 {
		return -1, errConstOverflow{}
	}
	p.constants = append(p.constants, v)
	return len(p.constants) - 1, nil
}

func (p *prog) count() int { return len(p.code) }

type errConstOverflow struct{}

func (errConstOverflow) Error() string { return "const overflow" }
