package bcl

import (
	"bufio"
	"errors"
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

// prog dump format:
//
// 2B: bytecode magic, then version: 1B: major, 1B: minor
// uvarint + n bytes: prog name
// uvarint + n bytes: code
// uvarint + n values: constants
// uvarint + n uvarints: positions
// uvarint + n uvarints: linepos (lfs)

const (
	bytecodeMagic       = "\xFC\x6C"
	bytecodeMajor uint8 = 2
	bytecodeMinor uint8 = 0
)

// Dump writes bytecode to the dest writer; on error, dest might hold partial data.
func (prog *Prog) Dump(dest io.Writer) error {
	if len(prog.code) > maxCodeLen {
		return fmt.Errorf("code too large: %d bytes, max %d", len(prog.code), maxCodeLen)
	}

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
	for i, v := range prog.constants {
		// all but string can fit in a fixed buffer
		if s, ok := v.(string); ok {
			if len(s) > maxStringLen {
				return fmt.Errorf(
					"constant[%d]: string too long: %d bytes, max %d", i, len(s), maxStringLen,
				)
			}
			wantedLen := 1 + 9 + len(s) // type + max varint for size + string
			if wantedLen > len(p) {
				p = make([]byte, wantedLen)
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
	var m uint64

	_, err = readPart(r, b[:2])
	if err != nil {
		return fmt.Errorf("magic header: %w", err)
	}
	if string(b[:2]) != bytecodeMagic {
		return fmt.Errorf("invalid magic header")
	}
	_, err = readPart(r, b[:2])
	if err != nil {
		return fmt.Errorf("bcode version: %w", err)
	}
	if b[0] != bytecodeMajor {
		return fmt.Errorf("invalid bcode major version: have %d.x, want %d.x", b[0], bytecodeMajor)
	}
	if b[1] > bytecodeMinor {
		return fmt.Errorf("invalid bcode minor version: have %d.%d, want <=%d.%d",
			b[0], b[1], b[0], bytecodeMinor,
		)
	}

	prog.name, err = stringFromBuf(r)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("code size: %w", err)
	}
	if m > maxCodeLen {
		return fmt.Errorf("code size: %d bytes, max %d", m, maxCodeLen)
	}
	prog.code = make([]byte, m)
	_, err = readPart(r, prog.code)
	if err != nil {
		return fmt.Errorf("code too short: %w", err)
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("constants size: %w", err)
	}
	// each constant is referenced at least once in the code
	if m > uint64(len(prog.code)) {
		return fmt.Errorf("constants size: %d exceeds code size %d", m, len(prog.code))
	}
	prog.constants = make([]value, m)
	for i := range prog.constants {
		prog.constants[i], err = valueFromBuf(r)
		if err != nil {
			return fmt.Errorf("constant[%d]: %w", i, err)
		}
	}

	m, err = uvarintFromBuf(r)
	if err != nil {
		return fmt.Errorf("positions size: %w", err)
	}
	// each code byte has its position
	if m != uint64(len(prog.code)) {
		return fmt.Errorf("positions size: %d, want code size %d", m, len(prog.code))
	}
	prog.positions = make([]int, m)
	for i := range prog.positions {
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

	_, err = r.ReadByte()
	switch err {
	case io.EOF:
		return nil
	case nil:
		return errors.New("unexpected trailing data")
	default:
		return err
	}
}
