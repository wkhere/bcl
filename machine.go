package bcl

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type vmConfig struct{ trace bool }

func execute(p *Prog, cf vmConfig) ([]Block, Binding, execStats, error) {
	vm := &vm{
		output: p.output,
		trace:  cf.trace,
		prog:   p,
		pc:     0,
	}
	err := vm.run()

	vm.stats.pcFinal = vm.pc
	return vm.result, vm.binding, vm.stats, err
}

type vm struct {
	prog  *Prog
	pc    int
	tos   int
	stack [stackSize]value

	blockTos   int
	blockStack [blockStackSize]Block

	output io.Writer
	log    io.Writer
	trace  bool

	result       []Block
	binding      Binding
	umbrellaOpen bool

	stats execStats
}

const (
	stackSize      = 1024
	blockStackSize = 16

	bindMaxNBlocks = 64
)

type execStats struct {
	tosMax      int
	blockTosMax int
	opsRead     int
	pcFinal     int
}

func (vm *vm) run() error {

	readByte := func() (b byte) {
		b = vm.prog.code[vm.pc]
		vm.pc++
		return b
	}
	readOp := func() (o opcode) {
		o = opcode(readByte())
		vm.stats.opsRead++
		return o
	}
	readU16 := func() int {
		x := u16FromBytes(vm.prog.code[vm.pc : vm.pc+2])
		vm.pc += 2
		return int(x)
	}
	readUvarint := func() int {
		x, n := uvarintFromBytes(vm.prog.code[vm.pc:])
		vm.pc += n
		return int(x)
	}
	readConst := func() value {
		return vm.prog.constants[readUvarint()]
	}

	push := func(v value) {
		vm.stack[vm.tos] = v
		vm.tos++
		vm.stats.tosMax = max(vm.stats.tosMax, vm.tos)
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

	getField := func(name string) (v value, ok bool) {
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
	setField := func(name string, v value) {
		vm.blockStack[vm.blockTos-1].Fields[name] = v
	}

	for {
		if vm.trace {
			printStack(vm.output, vm.stack[:vm.tos])
			vm.prog.disasmInstr(vm.pc)
		}

		switch instr := readOp(); instr {

		case opCONST:
			// ( -- x )
			push(readConst())

		case opZERO:
			// ( -- 0 )
			push(0)

		case opONE:
			// ( -- 1 )
			push(1)

		case opTRUE:
			// ( -- true )
			push(true)

		case opFALSE:
			// ( -- false )
			push(false)

		case opNIL:
			// ( -- nil )
			push(nil)

		case opEQ, opLT, opGT, opADD, opSUB, opMUL, opDIV:
			// ( a b -- c )
			switch {
			case isNumber(peek(1)) && isNumber(peek(0)):
				if b := peek(0); instr == opDIV && isInt(b) && b == 0 {
					return vm.runtimeError("division by int zero")
				}
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

			case instr == opADD && isString(peek(1)) && peek(0) == nil:
				pop()

			case instr == opMUL && isString(peek(1)) && isInt(peek(0)):
				b, a := pop().(int), pop().(string)
				push(strings.Repeat(a, b))

			case instr == opEQ:
				b, a := pop(), pop()
				push(a == b)

			default:
				return vm.runtimeError(
					"%s: invalid types: %s, %s", instr, vtype(peek(1)), vtype(peek(0)),
				)
			}

		case opNEG:
			// ( a -- b )
			if !isNumber(peek(0)) {
				return vm.runtimeError("NEG: invalid type: %s, expected number", vtype(peek(0)))
			}
			set(unopNumeric(instr, peek(0)))

		case opUNPLUS:
			// ( a -- a )
			if !isNumber(peek(0)) {
				return vm.runtimeError("UNPLUS: invalid type: %s, expected number", vtype(peek(0)))
			}
			// do nothing

		case opNOT:
			// ( a -- b )
			set(isFalsey(peek(0)))

		case opJUMP:
			// ( -- )
			vm.pc += readU16()

		case opLOOP:
			// ( -- )
			vm.pc -= readU16()

		case opJFALSE:
			// ( a -- a )
			jump := readU16()
			if isFalsey(peek(0)) {
				vm.pc += jump
			}

		case opPOP:
			// ( a -- )
			vm.tos--

		case opPOPN:
			// ( a1 ..aN -- )
			vm.tos -= int(readUvarint())

		case opPRINT:
			// ( a -- )
			fmt.Fprintln(vm.output, pop())

		case opGETLOCAL:
			// ( -- x )
			slot := readUvarint()
			push(vm.stack[slot])

		case opSETLOCAL:
			// ( x -- x )
			slot := readUvarint()
			vm.stack[slot] = peek(0)

		case opDEFBLOCK:
			// ( -- )
			blk := Block{
				Type:   readConst().(string),
				Name:   readConst().(string),
				Fields: map[string]any{},
			}
			vm.blockStack[vm.blockTos] = blk
			vm.blockTos++
			vm.stats.blockTosMax = max(vm.stats.blockTosMax, vm.blockTos)

		case opENDBLOCK:
			// ( -- )
			vm.blockTos--
			i := vm.blockTos
			if i > 0 {
				var (
					child  = &vm.blockStack[i]
					parent = &vm.blockStack[i-1]
					k      = child.key()
				)
				// note: good to have the following safety check, although
				// with the current syntax and with the key=type.name,
				// it is impossible to trigger it
				if _, ok := parent.Fields[k]; ok {
					return vm.runtimeError("child %s duplicate at parent", k)
				}
				parent.Fields[k] = *child

			} else {
				// todo: uniqueness check
				vm.result = append(vm.result, vm.blockStack[0])
			}

		case opGETFIELD:
			// ( -- x )
			name := readConst().(string)
			v, ok := getField(name)
			if !ok {
				return vm.runtimeError("identifier '%s' not resolved as var or field", name)
			}
			push(v)

		case opSETFIELD:
			// ( x -- x )
			name := readConst().(string)
			setField(name, peek(0))

		case opDEFUBIND:
			// ( -- )
			if vm.umbrellaOpen {
				return vm.runtimeError("nested umbrella binding is not allowed")
			}
			if vm.binding != nil {
				vm.warning("repeated bind statement overrides previous one")
			}
			vm.binding = &UmbrellaBinding{}
			vm.umbrellaOpen = true

		case opENDUBIND:
			// ( -- )
			if !vm.umbrellaOpen {
				return vm.runtimeError("expected umbrella binding present")
			}
			vm.umbrellaOpen = false

		case opBIND:
			// ( -- )
			if vm.binding != nil && !vm.umbrellaOpen {
				vm.warning("repeated bind statement overrides previous one")
			}

			var (
				bindCode     = readByte()
				blockType    = readConst().(string)
				n            = readUvarint()
				subSelectors []value
			)
			if n > bindMaxNBlocks {
				return vm.runtimeError(
					"bind: too many selected blocks, type=%s, n=%d", blockType, n,
				)
			}
			subSelectors = make([]value, n)
			for i := range subSelectors {
				subSelectors[i] = readConst()
			}

			if err := vm.bind(bindCode, blockType, subSelectors); err != nil {
				return err
			}

		case opRET:
			// ( -- )
			if vm.tos != 0 {
				return fmt.Errorf("internal error: non-empty stack on prog end; tos=%d", vm.tos)
			}
			return nil

		case opNOP:
			// ( -- )
		}
	}
}

func (vm *vm) bind(bindCode byte, blockType string, subSelectors []value) error {
	var (
		selector = bindSelector(bindCode & 0x0F)
		target   = bindTarget(bindCode & 0xF0)
		blocks   = make([]Block, 0, 1)
		selected []Block
		binding  Binding
	)

	for _, b := range vm.result {
		if b.Type == blockType {
			blocks = append(blocks, b)
		}
	}

	switch {
	case len(blocks) == 0:
		return vm.runtimeError("bind: no blocks of type %s", blockType)

	case len(blocks) != 1 && selector == bindOne:
		return vm.runtimeError(
			"bind: found %d blocks of type %s but expected just 1", len(blocks), blockType,
		)

	case selector == bindNamedBlock:
		name := subSelectors[0].(string)
		var found bool
		for i := range blocks {
			if blocks[i].Name == name {
				selected = []Block{blocks[i]}
				found = true
				break
			}
		}
		if !found {
			return vm.runtimeError("bind: block %s:%q not found", blockType, name)
		}

	case selector == bindNamedBlocks:
		if len(subSelectors) == 0 {
			return vm.runtimeError("bind: empty selection of blocks, type=%s", blockType)
		}

		blockMap := make(map[string]*Block, len(blocks))

		for i := range blocks {
			b := &blocks[i]
			blockMap[b.Name] = b
		}
		for _, sel := range subSelectors {
			name := sel.(string)
			b, ok := blockMap[name]
			if !ok {
				return vm.runtimeError("bind: block %s:%q not found", blockType, name)
			}
			selected = append(selected, *b)
		}
	}

	switch {
	case target == bindStruct && selector == bindOne:
		fallthrough

	case target == bindStruct && selector == bindFirst:
		binding = StructBinding{Value: blocks[0]}

	case target == bindStruct && selector == bindLast:
		binding = StructBinding{Value: blocks[len(blocks)-1]}

	case target == bindStruct && selector == bindNamedBlock:
		binding = StructBinding{Value: selected[0]}

	case target == bindSlice && selector == bindAll:
		binding = SliceBinding{Value: blocks}

	case target == bindSlice && selector == bindNamedBlocks:
		binding = SliceBinding{Value: selected}

	default:
		return vm.runtimeError("invalid bind target and selector: 0x%2x", bindCode)
	}

	if vm.umbrellaOpen {
		ub := vm.binding.(*UmbrellaBinding)
		ub.Parts = append(ub.Parts, binding)
	} else {
		vm.binding = binding
	}

	return nil
}

func (vm *vm) runtimeError(format string, a ...any) error {
	b := new(strings.Builder)
	pos := vm.prog.positions[vm.pc-1]
	fmt.Fprintf(b, "runtime error: line %s: ", vm.prog.linePos.format(pos))
	fmt.Fprintf(b, format, a...)
	return &runtimeErr{b.String()}
}

func (vm *vm) warning(format string, a ...any) {
	pos := vm.prog.positions[vm.pc-1]
	w := vm.prog.log
	fmt.Fprintf(w, "WARNING: line %s: ", vm.prog.linePos.format(pos))
	fmt.Fprintf(w, format+"\n", a...)
}

func (b *Block) key() string {
	if b.Name == "" {
		return b.Type
	}
	return b.Type + "." + b.Name
}

func printStack(w io.Writer, vv []value) {
	fmt.Fprintf(w, "             %d: ", len(vv))
	for _, v := range vv {
		fmt.Fprintf(w, "[ %v ]", v)
	}
	fmt.Fprintln(w)
}

type runtimeErr struct {
	msg string
}

func (e *runtimeErr) Error() string { return e.msg }
