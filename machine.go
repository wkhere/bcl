package bcl

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	stackSize      = 1024
	blockStackSize = 16
)

type vm struct {
	prog  *prog
	pc    int
	stack [stackSize]value
	tos   int

	trace bool

	blockStack  [blockStackSize]Block
	blockTos    int
	blockNextID int

	result []Block
}

func (vm *vm) execute(p *prog) error {
	vm.prog = p
	vm.pc = 0
	return vm.run()
}

func (vm *vm) run() error {

	readByte := func() (b byte) {
		b = vm.prog.code[vm.pc]
		vm.pc++
		return b
	}
	readU16 := func() int {
		x := u16FromBytes(vm.prog.code[vm.pc : vm.pc+2])
		vm.pc += 2
		return int(x)
	}

	readConst := func() value {
		return vm.prog.constants[readByte()]
	}

	push := func(v value) {
		vm.stack[vm.tos] = v
		vm.tos++
	}
	pop := func() value {
		vm.tos--
		return vm.stack[vm.tos]
	}
	peek := func(distance int) value {
		return vm.stack[vm.tos-1-distance]
	}
	set := func(v value) {
		vm.stack[vm.tos-1] = v
	}

	blockGet := func(name string) (v value, ok bool) {
		switch name {
		case "TYPE":
			return vm.blockStack[vm.blockTos-1].Type, true
		case "NAME":
			return vm.blockStack[vm.blockTos-1].Name, true
		}

		for i := vm.blockTos - 1; i >= 0; i-- {
			v, ok = vm.blockStack[i].Fields[name]
			if ok {
				return v, ok
			}
		}
		return
	}
	blockSet := func(name string, v value) {
		vm.blockStack[vm.blockTos-1].Fields[name] = v
	}

	for {
		if vm.trace {
			printStack(vm.stack[:vm.tos])
			vm.prog.disasmInstr(vm.pc)
		}

		switch instr := opcode(readByte()); instr {

		case opCONST:
			push(readConst())

		case opZERO:
			push(0)

		case opONE:
			push(1)

		case opTRUE:
			push(true)

		case opFALSE:
			push(false)

		case opNIL:
			push(nil)

		case opEQ, opLT, opGT, opADD, opSUB, opMUL, opDIV:
			switch {
			case isNumber(peek(1)) && isNumber(peek(0)):
				b, a := pop(), pop()
				push(binopNumeric(instr, a, b))

			case (instr == opLT || instr == opGT || instr == opADD) &&
				isString(peek(1)) && isString(peek(0)):
				b, a := pop().(string), pop().(string)
				push(binopString(instr, a, b))

			case instr == opADD && isString(peek(1)) && isInt(peek(0)):
				b, a := pop().(int), pop().(string)
				push(a + strconv.Itoa(b))

			case instr == opADD && isString(peek(1)) && isFloat(peek(0)):
				b, a := pop().(float64), pop().(string)
				push(a + strconv.FormatFloat(b, 'f', -1, 64))

			case instr == opMUL && isString(peek(1)) && isInt(peek(0)):
				b, a := pop().(int), pop().(string)
				push(strings.Repeat(a, b))

			case instr == opEQ:
				b, a := pop(), pop()
				push(a == b)

			default:
				return vm.runtimeError(
					"%s: invalid types: %s, %s",
					instr, vtype(peek(1)), vtype(peek(0)),
				)
			}

		case opNEG:
			if !isNumber(peek(0)) {
				return vm.runtimeError("NEG: invalid type: %s, expected number", vtype(peek(0)))
			}
			set(unopNumeric(instr, peek(0)))

		case opNOT:
			set(isFalsey(peek(0)))

		case opJUMP:
			vm.pc += readU16()

		case opLOOP:
			vm.pc -= readU16()

		case opJFALSE:
			jump := readU16()
			if isFalsey(peek(0)) {
				vm.pc += jump
			}

		case opPOP:
			vm.tos--

		case opPOPN:
			vm.tos -= int(readByte())

		case opPRINT:
			fmt.Println(pop())

		case opGETLOCAL:
			slot := readByte()
			push(vm.stack[slot])

		case opSETLOCAL:
			slot := readByte()
			vm.stack[slot] = peek(0)

		case opDEFBLOCK:
			blk := Block{
				Type:   readConst().(string),
				Name:   readConst().(string),
				Fields: map[string]any{},
			}
			vm.blockStack[vm.blockTos] = blk
			vm.blockTos++

		case opENDBLOCK:
			vm.blockTos--
			i := vm.blockTos
			if i > 0 {
				var (
					child  = &vm.blockStack[i]
					parent = &vm.blockStack[i-1]
					k      = child.key(vm)
				)
				if _, ok := parent.Fields[k]; ok {
					return vm.runtimeError("child %s duplicate at parent", k)
				}
				parent.Fields[k] = *child

			} else {
				// todo: uniqueness check
				vm.result = append(vm.result, vm.blockStack[0])
			}

		case opGETFIELD:
			name := readConst().(string)
			v, ok := blockGet(name)
			if !ok {
				return vm.runtimeError("identifier %q not resolved as var or field", name)
			}
			push(v)

		case opSETFIELD:
			name := readConst().(string)
			blockSet(name, peek(0))

		case opRET:
			if vm.tos != 0 {
				return fmt.Errorf("internal error: non-empty stack on prog end; tos=%d", vm.tos)
			}
			return nil

		case opNOP:
		}
	}
}

func (vm *vm) runtimeError(format string, a ...any) error {
	b := new(strings.Builder)
	pos := vm.prog.positions[vm.pc-1]
	fmt.Fprintf(b, "runtime error: line %s: ", vm.prog.lineFmt(pos))
	fmt.Fprintf(b, format, a...)
	return &runtimeErr{b.String()}
}

func (b *Block) key(vm *vm) string {
	id := b.Name
	if id == "" {
		vm.blockNextID++
		id = "#" + strconv.Itoa(vm.blockNextID)
	}
	return b.Type + "." + id
}

func printStack(vv []value) {
	fmt.Print("             ")
	for _, v := range vv {
		fmt.Printf("[ %v ]", v)
	}
	fmt.Println()
}

type runtimeErr struct {
	msg string
}

func (e *runtimeErr) Error() string { return e.msg }
