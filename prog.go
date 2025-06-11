package bcl

import (
	"bufio"
	"fmt"
	"io"
)

type Prog struct {
	name      string
	code      []byte
	constants []value
	positions []int
	linePos   *lineCalc

	output, log io.Writer
}

func newProg(name string, w writers) *Prog {
	return &Prog{
		name:   name,
		output: w.outw,
		log:    w.logw,
	}
}

func (p *Prog) initForParse() {
	const (
		codeInitCap      = 64
		constantsInitCap = 8
	)
	p.code = make([]byte, 0, codeInitCap)
	p.positions = make([]int, 0, codeInitCap)
	p.constants = make([]value, 0, constantsInitCap)
}

func (p *Prog) write(b byte, pos int) {
	p.code = append(p.code, b)
	p.positions = append(p.positions, pos)
}

func (p *Prog) addConst(v value) (idx int) {
	p.constants = append(p.constants, v)
	return len(p.constants) - 1
}

func (p *Prog) count() int { return len(p.code) }

// prog dump format, version 1.2
//
// 2B: bytecode magic, then version: 1B: major, 1B: minor
// uvarint + n bytes: prog name
// uvarint + n bytes: code
// uvarint + n values: constants
// uvarint + n uvarints: positions
// uvarint + n uvarints: linepos (lfs)

const (
	bytecodeMagic       = "\xFC\x6C"
	bytecodeMajor uint8 = 1
	bytecodeMinor uint8 = 2
)

func (prog *Prog) Dump(dest io.Writer) error {
	w := bufio.NewWriterSize(dest, 4096)
	w.Write(append([]byte(bytecodeMagic), bytecodeMajor, bytecodeMinor))

	var b [96]byte
	var p = b[:]
	var n int

	n = uvarintToBytes(p, uint64(len(prog.name)))
	w.Write(p[:n])
	w.Write([]byte(prog.name))

	n = uvarintToBytes(p, uint64(len(prog.code)))
	w.Write(p[:n])
	w.Write(prog.code)

	n = uvarintToBytes(p, uint64(len(prog.constants)))
	w.Write(p[:n])
	for _, v := range prog.constants {
		// all but string can fit in a fixed buffer
		if s, ok := v.(string); ok {
			if 2+len(s) > len(p) {
				p = make([]byte, 2+len(s))
			}
		}
		n = valueToBytes(p, v)
		w.Write(p[:n])
	}

	n = uvarintToBytes(p, uint64(len(prog.positions)))
	w.Write(p[:n])
	for _, x := range prog.positions {
		n = uvarintToBytes(p, uint64(x))
		w.Write(p[:n])
	}

	n = uvarintToBytes(p, uint64(len(prog.linePos.lfs)))
	w.Write(p[:n])
	for _, x := range prog.linePos.lfs {
		n = uvarintToBytes(p, uint64(x))
		w.Write(p[:n])
	}

	return w.Flush()
}

func (prog *Prog) Load(src io.Reader) (err error) {
	r := bufio.NewReaderSize(src, 4096)

	var b [2]byte
	var n int
	var m uint64

	n, _ = r.Read(b[:2])
	if n != 2 {
		return fmt.Errorf("missing magic header")
	}
	if string(b[:2]) != bytecodeMagic {
		return fmt.Errorf("invalid magic header")
	}
	n, _ = r.Read(b[:2])
	if n != 2 {
		return fmt.Errorf("missing bcode major/minor version")
	}
	if b[0] != bytecodeMajor {
		return fmt.Errorf("invalid bcode major version")
	}
	if b[1] > bytecodeMinor {
		return fmt.Errorf("invalid bcode minor version")
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("name size: %w", err)
	}
	p, err := r.Peek(int(m))
	if err != nil {
		return fmt.Errorf("name too short: %w", err)
	}
	r.Discard(int(m))
	prog.name = string(p)

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("code size: %w", err)
	}
	prog.code = make([]byte, m)
	n, err = io.ReadFull(r, prog.code)
	if n < int(m) {
		if err != nil {
			return fmt.Errorf("code too short: %w", err)
		}
		return fmt.Errorf("code too short")
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("constants size: %w", err)
	}
	prog.constants = make([]value, int(m))
	for i := 0; i < int(m); i++ {
		prog.constants[i], err = valueFromBuf(r)
		if err != nil {
			return fmt.Errorf("constant[%d]: %w", i, err)
		}
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("positions size: %w", err)
	}
	prog.positions = make([]int, int(m))
	for i := 0; i < int(m); i++ {
		x, err := uvarintFromBuf(r)
		if err != nil {
			return fmt.Errorf("position[%d]: %w", i, err)
		}
		prog.positions[i] = int(x)
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("lfs size: %w", err)
	}
	prog.linePos = &lineCalc{lfs: make([]int, int(m))}
	for i := 0; i < int(m); i++ {
		x, err := uvarintFromBuf(r)
		if err != nil {
			return fmt.Errorf("lfs[%d]: %w", i, err)
		}
		prog.linePos.lfs[i] = int(x)
	}

	_, err = r.Read(b[:1])
	if err == io.EOF {
		return nil
	}
	return err
}
