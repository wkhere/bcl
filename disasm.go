package bcl

import "fmt"

func (p *prog) disasm() {
	if p.name != "" {
		fmt.Println("==", p.name, "==")
	}

	for offset := 0; offset < len(p.code); {
		offset = p.disasmInstr(offset)
	}
}

func (p *prog) disasmInstr(offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && p.positions[offset] == p.positions[offset-1] {
		fmt.Printf("     |  ")
	} else {
		fmt.Printf("%6s  ", p.lineFmt(p.positions[offset]))
	}

	instr := opcode(p.code[offset])
	switch instr {
	case
		opNOP, opRET, opPRINT, opPOP,
		opENDBLOCK,
		opNIL, opZERO, opONE, opTRUE, opFALSE,
		opEQ, opLT, opGT,
		opADD, opSUB, opMUL, opDIV, opNEG, opNOT:
		return simpleInstr(instr, offset)

	case opCONST, opGETFIELD, opSETFIELD:
		return constInstr(instr, p, offset)

	case opGETLOCAL, opSETLOCAL, opPOPN:
		return byteargInstr(instr, p, offset)

	case opDEFBLOCK:
		return blockInstr(instr, p, offset)

	case opJUMP, opJFALSE:
		return jumpInstr(instr, +1, p, offset)
	case opLOOP:
		return jumpInstr(instr, -1, p, offset)

	default:
		fmt.Println("unknown opcode", instr)
		return offset + 1
	}
}

func simpleInstr(o opcode, offset int) int {
	fmt.Println(o)
	return offset + 1
}

func byteargInstr(o opcode, p *prog, offset int) int {
	arg := p.code[offset+1]
	fmt.Printf("%-10s %4d\n", o, arg)
	return offset + 2
}

func twobyteargInstr(o opcode, p *prog, offset int) int {
	arg := u16FromBytes(p.code[offset+1 : offset+3])
	fmt.Printf("%-10s %4d\n", o, arg)
	return offset + 3
}

func constInstr(o opcode, p *prog, offset int) int {
	constIdx := p.code[offset+1]
	fmt.Printf("%-10s %4d '%v'\n", o, constIdx, p.constants[constIdx])
	return offset + 2
}

func blockInstr(o opcode, p *prog, offset int) int {
	typeIdx := p.code[offset+1]
	nameIdx := p.code[offset+2]
	fmt.Printf(
		"%-10s %4d '%v'  %4d '%v'\n",
		o, typeIdx, p.constants[typeIdx], nameIdx, p.constants[nameIdx],
	)
	return offset + 3
}

func jumpInstr(o opcode, sign int, p *prog, offset int) int {
	jump := u16FromBytes(p.code[offset+1 : offset+3])
	fmt.Printf("%-10s -> %04d\n", o, offset+3+sign*int(jump))
	return offset + 3
}
