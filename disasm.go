package bcl

import (
	"fmt"
	"io"
)

func (p *Prog) disasm() {
	if p.name != "" {
		fmt.Fprintln(p.output, "==", p.name, "==")
	}

	for offset := 0; offset < len(p.code); {
		offset = p.disasmInstr(offset)
	}
}

func (p *Prog) disasmInstr(offset int) int {
	fmt.Fprintf(p.output, "%04d ", offset)
	if offset > 0 && p.positions[offset] == p.positions[offset-1] {
		fmt.Fprintf(p.output, "     |  ")
	} else {
		fmt.Fprintf(p.output, "%6s  ", p.linePos.format(p.positions[offset]))
	}

	instr := opcode(p.code[offset])
	switch instr {
	case
		opNOP, opRET, opPRINT, opPOP,
		opENDBLOCK,
		opNIL, opZERO, opONE, opTRUE, opFALSE,
		opEQ, opLT, opGT,
		opADD, opSUB, opMUL, opDIV, opNEG, opNOT:
		return simpleInstr(p.output, instr, offset)

	case opCONST, opGETFIELD, opSETFIELD:
		return constInstr(p.output, instr, p, offset)

	case opGETLOCAL, opSETLOCAL, opPOPN:
		return varbyteargInstr(p.output, instr, p, offset)

	case opDEFBLOCK:
		return blockInstr(p.output, instr, p, offset)

	case opJUMP, opJFALSE:
		return jumpInstr(p.output, instr, +1, p, offset)
	case opLOOP:
		return jumpInstr(p.output, instr, -1, p, offset)

	default:
		fmt.Fprintln(p.output, "unknown opcode", instr)
		return offset + 1
	}
}

func simpleInstr(w io.Writer, o opcode, offset int) int {
	fmt.Fprintln(w, o)
	return offset + 1
}

func byteargInstr(w io.Writer, o opcode, p *Prog, offset int) int {
	arg := p.code[offset+1]
	fmt.Fprintf(w, "%-10s %4d\n", o, arg)
	return offset + 2
}

func varbyteargInstr(w io.Writer, o opcode, p *Prog, offset int) int {
	arg, n := uvarintFromBytes(p.code[offset+1:])
	fmt.Fprintf(w, "%-10s %4d\n", o, arg)
	return offset + 1 + n
}

func constInstr(w io.Writer, o opcode, p *Prog, offset int) int {
	idx, n := uvarintFromBytes(p.code[offset+1:])
	fmt.Fprintf(w, "%-10s %4d '%v'\n", o, idx, p.constants[idx])
	return offset + 1 + n
}

func blockInstr(w io.Writer, o opcode, p *Prog, offset int) int {
	typeIdx, n1 := uvarintFromBytes(p.code[offset+1:])
	nameIdx, n2 := uvarintFromBytes(p.code[offset+1+n1:])
	fmt.Fprintf(w,
		"%-10s %4d '%v'\t%4d '%v'\n",
		o, typeIdx, p.constants[typeIdx], nameIdx, p.constants[nameIdx],
	)
	return offset + 1 + n1 + n2
}

func jumpInstr(w io.Writer, o opcode, sign int, p *Prog, offset int) int {
	jump := u16FromBytes(p.code[offset+1:])
	fmt.Fprintf(w, "%-10s %4d -> %04d\n",
		o, jump, offset+1+jumpByteLength+sign*int(jump),
	)
	return offset + 1 + jumpByteLength
}
