package bcl

// Lexer based on "Lexical Scanning in Go" by Rob Pike.

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// API

// newLexer creates new nexer and runs its loop.
func newLexer(inputs <-chan string, linePosUpdater func(string, int)) *lexer {
	l := &lexer{
		inputs: inputs,
		lpUpd:  linePosUpdater,
		tokens: make(chan token, tokensBufSize),
	}
	go l.run()
	return l
}

// The parser calls nextToken to get the actual token.
func (l *lexer) nextToken() (_ token, ok bool) {
	token, ok := <-l.tokens
	return token, ok
}

// engine

const tokensBufSize = 10

type lexer struct {
	inputs     <-chan string
	input      string
	lpUpd      func(string, int)
	start, pos int
	posShift   int
	width      int
	tokens     chan token
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{
		typ: t,
		val: string(l.input[l.start:l.pos]),
		pos: l.pos + l.posShift,
	}
	l.start = l.pos
}

func (l *lexer) emitError(format string, args ...any) {
	l.tokens <- token{
		typ: tERR,
		err: fmt.Errorf(format, args...),
		pos: l.pos + l.posShift,
	}
}

// input-consuming primitives

const (
	eof rune = -1
)

// next gets the next rune from the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		s, ok := <-l.inputs
		if !ok {
			if l.pos == l.start {
				l.width = 0
				return eof
			}
			// continue with leftover + s
		}
		l.input = l.input[l.start:l.pos] + s
		l.posShift += l.start
		l.lpUpd(s, l.posShift+l.pos-l.start)
		l.pos -= l.start
		l.start = 0
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	if l.width == 0 {
		return eof
	}
	l.pos += l.width
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// input-consuming helpers

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) (accepted bool) {
	for strings.ContainsRune(valid, l.next()) {
		accepted = true
	}
	l.backup()
	return accepted
}

// acceptRunFunc consumes a run of runes satisfying the predicate.
func (l *lexer) acceptRunFunc(pred func(rune) bool) {
	for pred(l.next()) {
	}
	l.backup()
}

// // skipUntil consumes runes until a predidate is satisfied.
// func (l *lexer) skipUntil(pred func(rune) bool) {
// 	for {
// 		if c := l.next(); c == eof || pred(c) {
// 			break
// 		}
// 	}
// 	l.backup()
// }

// rune predicates

// func isStrictSpace(r rune) bool {
// 	return r == ' ' || r == '\t'
// }

func isEol(r rune) bool {
	return r == '\n' || r == '\r'
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\v', '\f', '\n', '\r', 0x85, 0xA0:
		return true
	}
	return false
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlpha(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isAlphaNum(r rune) bool {
	return isAlpha(r) || isDigit(r)
}

// state finalizers

func (l *lexer) fail(format string, args ...any) stateFn {
	l.emitError(format, args...)
	l.ignore()
	l.emit(tFAIL)
	return nil
}

// state functions and related data

var keywords = map[string]tokenType{
	"var":   tVAR,
	"def":   tDEF,
	"eval":  tEVAL,
	"print": tPRINT,
	"true":  tTRUE,
	"false": tFALSE,
	"nil":   tNIL,
	"not":   tNOT,
	"and":   tAND,
	"or":    tOR,
}

type twoRuneMatch struct {
	r2  rune
	typ tokenType
}

var twoRuneTokens = map[rune]twoRuneMatch{
	'=': {'=', tEE},
	'!': {'=', tBE},
	'<': {'=', tLE},
	'>': {'=', tGE},
}

var oneRuneTokens = map[rune]tokenType{
	'=': tEQ,
	'{': tLCURLY,
	'}': tRCURLY,
	'(': tLPAREN,
	')': tRPAREN,
	//'[': tLBRACKET,
	//']': tRBRACKET,
	'<': tLT,
	'>': tGT,
	'+': tPLUS,
	'-': tMINUS,
	'*': tSTAR,
	'/': tSLASH,
	';': tSEMICOLON,
}

const (
	digits      = "0123456789"
	hexdigits   = digits + "abcdefABCDEF"
	lineComment = '#'
)

func lexStart(l *lexer) stateFn {

	r := l.next()
	r2m, r2ok := twoRuneTokens[r]
	r1t, r1ok := oneRuneTokens[r]

	switch {
	case r == eof:
		l.emit(tEOF)
		return nil
	case r2ok:
		r2 := l.next()
		if r2 == r2m.r2 {
			l.emit(r2m.typ)
			return lexStart
		}
		l.backup()
		if !r1ok {
			return l.fail(
				"expected char %q to start token %q", r,
				fmt.Sprintf("%c%c", r, r2m.r2),
			)
		}
		fallthrough
	case r1ok:
		l.emit(r1t)
		return lexStart
	case isSpace(r):
		return lexSpace
	case r == lineComment:
		return lexLineComment
	case r == '"':
		return lexQuote
	case isAlpha(r) || r == '_':
		return lexKeywordOrIdent
	case isDigit(r):
		return lexNumber
	default:
		return l.fail("unknown char %#U", r)
	}
}

func lexSpace(l *lexer) stateFn {
	l.acceptRunFunc(isSpace)
	l.ignore()
	return lexStart
}

func lexLineComment(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case isEol(r), r == eof:
			l.backup()
			l.ignore()
			return lexStart
		}
	}
}

func lexKeywordOrIdent(l *lexer) stateFn {
loop:
	for {
		switch r := l.next(); {
		case isAlphaNum(r) || r == '_':
			//eat.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			key, isKey := keywords[word]
			switch {
			case isKey:
				l.emit(key)
			default:
				l.emit(tIDENT)
			}
			break loop
		}
	}
	return lexStart
}

func lexNumber(l *lexer) stateFn {
	l.backup()
	if l.accept("0") && l.accept("xX") {
		return lexHex
	}
	l.acceptRun(digits)
	if r := l.peek(); r == '.' || r == 'e' || r == 'E' {
		return lexFloat
	}
	l.emit(tINT)
	return lexStart
}

func lexHex(l *lexer) stateFn {
	l.acceptRun(hexdigits)
	// omitting hex floats
	l.emit(tINT)
	return lexStart
}

func lexFloat(l *lexer) stateFn {
	if l.accept(".") {
		ok := l.acceptRun(digits)
		if !ok {
			return l.fail("need more digits after a dot")
		}
	}
	if l.accept("eE") {
		l.accept("+-")
		ok := l.acceptRun(digits)
		if !ok {
			return l.fail("need more digits for an exponent")
		}
	}
	l.emit(tFLOAT)
	return lexStart
}

func lexQuote(l *lexer) stateFn {
loop:
	for {
		switch r := l.next(); r {
		case '\\':
			if r = l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.fail("unterminated quoted string")
		case '"':
			break loop
		}
	}
	l.emit(tSTR)
	return lexStart
}
