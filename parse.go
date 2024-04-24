package bcl

import "strconv"

func parse(input string, cf config) (*prog, error) {
	p := &parser{
		lexer:   newLexer(input),
		prog:    newProg("input"),
		lineFmt: newLineCalc(input).format,

		identRefs: make(map[string]int, 8),
		// identRefs are for reusing block types & fields and selected consts

		scope: new(scopeCompiler),
	}

	p.advance()

	for !p.match(tEOF) {
		decl(p)

		p.match(tSEMICOLON) // no check as it is optional
	}

	if p.hadError {
		return p.prog, errCombined{"parse"}
	}
	p.end()

	if cf.disasm {
		p.prog.disasm()
	}

	return p.prog, nil
}

type parser struct {
	lexer   *lexer
	prog    *prog
	lineFmt func(pos int) string

	prev, current token
	hadError      bool
	panicMode     bool

	identRefs map[string]int // map ident names to const indices

	scope *scopeCompiler
}

type scopeCompiler struct {
	locals     [localsMaxSize]local
	localCount int
	depth      int
}

// localsMaxSize limit is bound by 1B index in the bytecode,
// similar to const indices.
// The solution is to allow 2-byte indices in the bytecode, possibly with
// extra opcodes.
const localsMaxSize = 256

type local struct {
	name  string
	depth int
}

func decl(p *parser) {
	if p.match(tVAR) {
		varDecl(p)
	} else {
		stmt(p)
	}

	if p.panicMode && p.scope.depth == 0 {
		p.sync()
	}
}

func varDecl(p *parser) {
	p.consume(tIDENT, "expected variable name")
	if p.panicMode {
		return
	}

	p.declVar()

	if p.match(tEQ) {
		expr(p)
	} else {
		p.emitOp(opNIL)
	}

	p.defVar()
}

func stmt(p *parser) {
	if p.match(tPRINT) {
		printStmt(p)
	} else if p.match(tEVAL) {
		exprStmt(p)
	} else if p.match(tDEF) {
		blockStmt(p)
	} else if p.scope.depth > 0 {
		expr(p)
	} else {
		p.errorAtCurrent("expected statement")
	}
}

func blockStmt(p *parser) {
	p.consume(tIDENT, "expected block type")
	if p.panicMode {
		return
	}

	blockType := p.prev.val

	var blockName string
	if p.match(tSTR) {
		blockName, _ = strconv.Unquote(p.prev.val)
	}

	p.consume(tLCURLY, "expected '{'")

	ok := p.defBlock(p.identConst(blockType), p.makeConst(blockName))
	if ok {
		defer p.endBlock()
	}
	// actually what to do if block couldn't be allocated??

	p.beginScope()
	defer p.endScope()

	for !p.check(tRCURLY) && !p.check(tEOF) {
		decl(p)
		if p.panicMode {
			p.advance()
		}

		p.match(tSEMICOLON) // optional
	}

	p.consume(tRCURLY, "expected '}'")
}

func printStmt(p *parser) {
	expr(p)
	p.emitOp(opPRINT)
}

func exprStmt(p *parser) {
	expr(p)
	p.emitOp(opPOP)
}

func expr(p *parser) {
	p.parsePrecedence(precAssign)
}

type parseFn func(*parser, bool)

type parseRule struct {
	prefix parseFn
	infix  parseFn
	prec   precedence
}
type precedence int

const (
	precNone precedence = iota
	precAssign
	precOr
	precAnd
	precNot
	precEq
	precCmp
	precTerm
	precFactor
	precUnary
	precCall
	precPrimary
)

var rules [tMAX]parseRule

func init() {
	rules = [...]parseRule{

		tLPAREN: {parens, nil, precNone},
		tRPAREN: {nil, nil, precNone},
		tLCURLY: {nil, nil, precNone},
		tRCURLY: {nil, nil, precNone},

		tEQ: {nil, nil, precNone},

		tMINUS: {unary, binary, precTerm},
		tPLUS:  {unary, binary, precTerm},
		tSLASH: {nil, binary, precFactor},
		tSTAR:  {nil, binary, precFactor},

		tOR:  {nil, boolOr, precOr},
		tAND: {nil, boolAnd, precAnd},
		tNOT: {boolNot, nil, precNone},

		tBE: {nil, binary, precEq},
		tEE: {nil, binary, precEq},
		tGT: {nil, binary, precCmp},
		tGE: {nil, binary, precCmp},
		tLT: {nil, binary, precCmp},
		tLE: {nil, binary, precCmp},

		tIDENT: {identRef, nil, precNone},
		tSTR:   {stringLit, nil, precNone},
		tINT:   {intLit, nil, precNone},
		tFLOAT: {floatLit, nil, precNone},

		tFALSE: {boolLit, nil, precNone},
		tTRUE:  {boolLit, nil, precNone},

		tNIL: {nilLit, nil, precNone},

		tVAR: {nil, nil, precNone},

		tSEMICOLON: {nil, nil, precNone},

		tERR: {nil, nil, precNone},
		tEOF: {nil, nil, precNone},
	}
}

func getRule(t tokenType) parseRule { return rules[t] }

func identRef(p *parser, canAssign bool) {
	p.resolveIdent(p.prev.val, canAssign)
}

func parens(p *parser, _ bool) {
	expr(p)
	p.consume(tRPAREN, "expected ')' after expression")
}

func binary(p *parser, _ bool) {
	opType := p.prev.typ
	rule := getRule(opType)

	p.parsePrecedence(rule.prec + 1)

	switch opType {
	case tEE:
		p.emitOp(opEQ)
	case tBE:
		p.emitOps(opEQ, opNOT)
	case tLT:
		p.emitOp(opLT)
	case tLE:
		p.emitOps(opGT, opNOT)
	case tGT:
		p.emitOp(opGT)
	case tGE:
		p.emitOps(opLT, opNOT)

	case tPLUS:
		p.emitOp(opADD)
	case tMINUS:
		p.emitOp(opSUB)
	case tSTAR:
		p.emitOp(opMUL)
	case tSLASH:
		p.emitOp(opDIV)
	}
}

func boolAnd(p *parser, _ bool) {
	endJump := p.emitJump(opJFALSE)

	p.emitOp(opPOP)
	p.parsePrecedence(precAnd)
	// NOTE: *not incrementing* prececence means that this op is right-assoc.
	// It doesn't seem to hurt in bool algebra; the bytecode is even cleaner.
	// Increment if ever going back to typical left-associativity.

	p.patchJump(endJump)
}

func boolOr(p *parser, _ bool) {
	midJump := p.emitJump(opJFALSE)
	endJump := p.emitJump(opJUMP)

	p.patchJump(midJump)
	p.emitOp(opPOP)
	p.parsePrecedence(precOr)
	// See NOTE for boolAnd and not incrementing the precedence.

	p.patchJump(endJump)
}

func boolNot(p *parser, _ bool) {
	opType := p.prev.typ

	p.parsePrecedence(precNot)

	switch opType {
	case tNOT:
		p.emitOp(opNOT)
	}
}

func unary(p *parser, _ bool) {
	opType := p.prev.typ

	p.parsePrecedence(precUnary)

	switch opType {
	case tMINUS:
		p.emitOp(opNEG)
	}
	// note: tPLUS also leads here and is ignored
}

func intLit(p *parser, _ bool) {
	v, err := strconv.ParseInt(p.prev.val, 0, 0)
	if err != nil {
		panic(err)
	}
	switch v {
	case 0:
		p.emitOp(opZERO)
	case 1:
		p.emitOp(opONE)
	default:
		p.emitConst(int(v))
	}
}

func floatLit(p *parser, _ bool) {
	v, err := strconv.ParseFloat(p.prev.val, 64)
	if err != nil {
		panic(err)
	}
	p.emitConst(v)
}

func stringLit(p *parser, _ bool) {
	s, err := strconv.Unquote(p.prev.val)
	if err != nil {
		panic(err)
	}
	p.emitConst(s)
}

func boolLit(p *parser, _ bool) {
	switch p.prev.typ {
	case tTRUE:
		p.emitOp(opTRUE)
	case tFALSE:
		p.emitOp(opFALSE)
	}
}

func nilLit(p *parser, _ bool) {
	switch p.prev.typ {
	case tNIL:
		p.emitOp(opNIL)
	}
}

func (p *parser) advance() {
	p.prev = p.current

	for {
		p.current = p.lexer.nextToken()
		if p.current.typ != tERR {
			break
		}
		p.errorAtCurrent(p.current.err.Error())
	}
}

func (p *parser) consume(typ tokenType, errmsg string) {
	if p.current.typ == typ {
		p.advance()
		return
	}

	p.errorAtCurrent(errmsg)
}

func (p *parser) match(typ tokenType) bool {
	if !p.check(typ) {
		return false
	}
	p.advance()
	return true
}

func (p *parser) check(typ tokenType) bool {
	return p.current.typ == typ
}

func (p *parser) sync() {
	p.panicMode = false

	for p.current.typ != tEOF {
		switch p.current.typ {
		case tVAR, tDEF, tPRINT, tEVAL: // tokens delimiting a statement
			return
		}
		p.advance()
	}
}

func (p *parser) parsePrecedence(prec precedence) {
	p.advance()
	prefixRule := getRule(p.prev.typ).prefix
	if prefixRule == nil {
		p.error("expected expression")
		return
	}

	canAssign := prec <= precAssign
	prefixRule(p, canAssign)

	for prec <= getRule(p.current.typ).prec {
		p.advance()
		infixRule := getRule(p.prev.typ).infix
		infixRule(p, canAssign)
	}

	if canAssign && p.match(tEQ) {
		p.error("invalid assignment target")
	}
}

func (p *parser) end() {
	p.popN(p.scope.localCount)
	p.emitOp(opRET)
	p.prog.lineFmt = p.lineFmt
}

func (p *parser) beginScope() {
	p.scope.depth++
}

func (p *parser) endScope() {
	p.scope.depth--

	var popCount int
	for p.scope.localCount > 0 &&
		p.scope.locals[p.scope.localCount-1].depth > p.scope.depth {
		popCount++
		p.scope.localCount--
	}
	p.popN(popCount)
}

func (p *parser) popN(count int) {
	switch count {
	case 0:
	case 1:
		p.emitOp(opPOP)
	default:
		p.emitOpWithArg(opPOPN, byte(count))
	}
}

func (p *parser) declVar() {
	name := p.prev.val

	for i := p.scope.localCount - 1; i >= 0; i-- {
		local := &p.scope.locals[i]
		if local.depth != -1 && local.depth < p.scope.depth {
			break
		}

		if name == local.name {
			p.error("variable with this name already present in this scope")
		}
	}

	p.addLocal(name)
}

func (p *parser) addLocal(name string) {
	if p.scope.localCount == localsMaxSize {
		p.error("too many local variables")
		return
	}

	local := &p.scope.locals[p.scope.localCount]
	local.name = name
	local.depth = -1
	p.scope.localCount++
}

func (p *parser) markInitialized() {
	p.scope.locals[p.scope.localCount-1].depth = p.scope.depth
}

func (p *parser) defVar() {
	p.markInitialized()
}

func (p *parser) defBlock(typeIdx, nameIdx int) bool {
	if typeIdx < 0 || nameIdx < 0 {
		return false
	}
	// todo: handle const indices longer than 1B
	p.emitOpWithArgs(opDEFBLOCK, byte(typeIdx), byte(nameIdx))
	return true
}

func (p *parser) endBlock() {
	p.emitOp(opENDBLOCK)
}

func (p *parser) resolveIdent(name string, canAssign bool) {
	var setOp, getOp opcode
	var idx int

	idx = p.resolveLocal(p.scope, name)
	if idx >= 0 {
		setOp, getOp = opSETLOCAL, opGETLOCAL
	} else {
		if p.scope.depth == 0 {
			p.error("undefined variable")
			return
		}
		// when in block, there can be field used at runtime,
		// or "ident not resolved" runtime error
		idx = p.identConst(name)
		if idx < 0 {
			return
		}
		setOp, getOp = opSETFIELD, opGETFIELD
	}

	// todo: handle var indices longer than 1B
	if canAssign && p.match(tEQ) {
		expr(p)
		p.emitOpWithArg(setOp, byte(idx))
	} else {
		p.emitOpWithArg(getOp, byte(idx))
	}
}

func (p *parser) resolveLocal(scope *scopeCompiler, name string) int {
	for i := scope.localCount - 1; i >= 0; i-- {
		local := &scope.locals[i]
		if name == local.name {
			if local.depth == -1 {
				// `var x=x` will reach x from outer scope
				continue
			}
			return i
		}
	}
	return -1
}

func (p *parser) emitByte(b byte) {
	p.currentProg().write(b, p.prev.pos)
}

func (p *parser) emitBytes(bb ...byte) {
	prog := p.currentProg()
	for _, b := range bb {
		prog.write(b, p.prev.pos)
	}
}

func (p *parser) emitOp(op opcode) {
	prog := p.currentProg()
	prog.write(byte(op), p.prev.pos)
}

func (p *parser) emitOpWithArg(op opcode, b byte) {
	prog := p.currentProg()
	prog.write(byte(op), p.prev.pos)
	prog.write(b, p.prev.pos)
}

func (p *parser) emitOpWithArgs(op opcode, bb ...byte) {
	prog := p.currentProg()
	prog.write(byte(op), p.prev.pos)
	for _, b := range bb {
		prog.write(b, p.prev.pos)
	}
}

func (p *parser) emitOps(oo ...opcode) {
	prog := p.currentProg()
	for _, o := range oo {
		prog.write(byte(o), p.prev.pos)
	}
}

func (p *parser) emitConst(v value) {
	c := p.makeConst(v)
	if c < 0 {
		return
	}
	// todo: handle const indices longer than 1B
	p.emitOpWithArgs(opCONST, byte(c))
}

func (p *parser) emitJump(op opcode) int {
	p.emitOp(op)
	p.emitBytes(0xff, 0xff)
	return p.currentProg().count() - 2
}

func (p *parser) patchJump(offset int) {
	prog := p.currentProg()
	jump := prog.count() - offset - 2

	if jump > 65535 {
		p.error("jump too long")
		return
	}
	if jump < 0 {
		p.error("negative jump")
		return
	}

	u16ToBytes(prog.code[offset:offset+2], uint16(jump))
}

func (p *parser) identConst(name string) int {
	idx, ok := p.identRefs[name]
	if !ok {
		idx = p.makeConst(name)
		p.identRefs[name] = idx
	}
	return idx
}

func (p *parser) makeConst(v value) int {
	// in case of some values like empty str, cache like identConst:
	var cache bool
	if v == "" {
		idx, ok := p.identRefs[""]
		if ok {
			return idx
		}
		cache = true
	}

	constIdx, err := p.currentProg().addConst(v)
	if err != nil {
		p.error(err.Error())
		return -1
	}
	if cache {
		p.identRefs[v.(string)] = constIdx
	}
	return constIdx
}

func (p *parser) currentProg() *prog {
	// ...
	return p.prog
}

func (p *parser) errorAtCurrent(msg string) {
	p.errorAt(&p.current, msg)
}

func (p *parser) error(msg string) {
	p.errorAt(&p.prev, msg)
}

func (p *parser) errorAt(t *token, msg string) {
	p.panicMode = true

	log.Printf("line %s: error", p.lineFmt(t.pos))

	switch {
	case t.typ == tEOF:
		log.Print(" at end")
	case t.typ == tERR: // nop
	default:
		log.Printf(" at '%s'", t.val)
	}

	log.Printf(": %s\n", msg)
	p.hadError = true
}

type errCombined struct {
	source string
}

func (e errCombined) Error() string { return "combined errors from " + e.source }
