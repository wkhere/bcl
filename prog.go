package bcl

import "io"

type prog struct {
	name      string
	code      []byte
	positions []int
	constants []value
	linePos   *lineCalc

	output io.Writer
}

func newProg(name string, output io.Writer) *prog {
	const (
		codeCap      = 64
		constantsCap = 8
	)
	return &prog{
		name:      name,
		code:      make([]byte, 0, codeCap),
		positions: make([]int, 0, codeCap),
		constants: make([]value, 0, constantsCap),
		output:    output,
	}
}

func (p *prog) write(b byte, pos int) {
	p.code = append(p.code, b)
	p.positions = append(p.positions, pos)
}

func (p *prog) addConst(v value) (idx int) {
	p.constants = append(p.constants, v)
	return len(p.constants) - 1
}

func (p *prog) count() int { return len(p.code) }
