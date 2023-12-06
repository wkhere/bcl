package bcl

// Lexer based on "Lexical Scanning in Go" by Rob Pike.

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// API

type itemType int // yacc tokens

type item struct {
	typ itemType
	val string
	err error // either val or err is set
	pos int
}

// newLexer creates new nexer and runs its loop.
func newLexer(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item, itemsBufSize),
	}
	go l.run()
	return l
}

// The parser calls nextItem to get the actual item.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.prevItem = l.lastItem
	l.lastItem = item
	return item
}

// engine

const itemsBufSize = 10

type lexer struct {
	input      string
	start, pos int
	width      int
	items      chan item

	// interface with parser:
	prevItem item
	lastItem item
	err      error
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{
		typ: t,
		val: string(l.input[l.start:l.pos]),
		pos: l.pos,
	}
	l.start = l.pos
}

func (l *lexer) emitError(format string, args ...any) {
	l.items <- item{
		typ: tERR,
		err: fmt.Errorf(format, args...),
		pos: l.pos,
	}
}

// input-consuming primitives

const (
	eof rune = -1
)

// next gets the next rune from the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
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
	return unicode.IsSpace(r)
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r)
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isOneRuneToken(r rune) bool {
	return strings.ContainsRune(oneRuneTokens, r)
}

// state finalizers

func (l *lexer) errorf(format string, args ...any) stateFn {
	l.emitError(format, args...)
	return nil
}

// state functions and related data

var keywords = map[string]itemType{
	"var":   tVAR,
	"def":   tDEF,
	"true":  tTRUE,
	"false": tFALSE,
	"not":   tNOT,
	"and":   tAND,
	"or":    tOR,
}

type twoRuneMatch struct {
	r2  rune
	typ itemType
}

var twoRuneTokens = map[rune]twoRuneMatch{
	'=': {'=', tEQ},
	'!': {'=', tNE},
	'<': {'=', tLE},
	'>': {'=', tGE},
}

const oneRuneTokens = "={}+-*/()<>"

const (
	digits      = "0123456789"
	hexdigits   = digits + "abcdefABCDEF"
	lineComment = '#'
)

func lexStart(l *lexer) stateFn {

	r := l.next()
	r2m, r2ok := twoRuneTokens[r]

	switch {
	case r == eof:
		l.emit(tEOF)
		// ^^is it needed as a terminator in the yparser?
		// if not then l.ignore() it
		return nil
	case r2ok:
		return lexTwoRunes(r, r2m)
	case isOneRuneToken(r):
		l.emit(itemType(r))
		return lexStart
	case isSpace(r):
		return lexSpace
	case r == lineComment:
		return lexLineComment
	case r == '"':
		return lexQuote
	case isAlpha(r) || r == '_':
		return lexIdentifier
	case isDigit(r):
		return lexNumber
	default:
		return l.errorf("unknown char %#U", r)
	}
}

func lexTwoRunes(r1 rune, match twoRuneMatch) stateFn {
	return func(l *lexer) stateFn {
		r2 := l.next()
		if r2 == match.r2 {
			l.emit(match.typ)
			return lexStart
		}
		l.backup()
		if isOneRuneToken(r1) {
			l.emit(itemType(r1))
			return lexStart
		}
		return l.errorf(
			"expected char %q to start token %q",
			r1, fmt.Sprintf("%c%c", r1, match.r2),
		)
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

func lexIdentifier(l *lexer) stateFn {
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
			return l.errorf("need more digits after a dot")
		}
	}
	if l.accept("eE") {
		l.accept("+-")
		ok := l.acceptRun(digits)
		if !ok {
			return l.errorf("need more digits for an exponent")
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
			return l.errorf("unterminated quoted string")
		case '"':
			break loop
		}
	}
	l.emit(tSTR)
	return lexStart
}
