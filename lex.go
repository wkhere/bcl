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
	typ  itemType
	val  string
	err  error // either val or err is set
	line int
}

// newLexer creates new nexer and runs its loop.
func newLexer(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item, itemsBufSize),
		line:  1,
	}
	go l.run()
	return l
}

// The parser calls nextItem to get the actual item.
func (l *lexer) nextItem() item {
	item := <-l.items
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
	line       int
	// interface with parser:
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
		typ:  t,
		val:  string(l.input[l.start:l.pos]),
		line: l.line,
	}
	l.start = l.pos
}

func (l *lexer) emitError(format string, args ...any) {
	l.items <- item{
		typ:  ERR_LEX,
		err:  fmt.Errorf(format, args...),
		line: l.line,
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
	if r == '\n' {
		l.line++
	}
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// // peek returns but does not consume the next rune in the input.
// func (l *lexer) peek() rune {
// 	r := l.next()
// 	l.backup()
// 	return r
// }

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// input-consuming helpers

// // accept consumes the next rune if it's from the valid set.
// func (l *lexer) acceptOne(valid string) bool {
// 	if strings.ContainsRune(valid, l.next()) {
// 		return true
// 	}
// 	l.backup()
// 	return false
// }

// acceptRunFromSet consumes a run of runes from the valid set.
func (l *lexer) acceptRunFromSet(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// acceptRun consumes a run of runes satisfying the predicate.
func (l *lexer) acceptRun(pred func(rune) bool) {
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

const shortTokens = "={}+-*/()"

// func isStrictSpace(r rune) bool {
// 	return r == ' ' || r == '\t'
// }

func isEol(r rune) bool {
	return r == '\n' || r == '\r'
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isShortToken(r rune) bool {
	return strings.ContainsRune(shortTokens, r)
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

// state finalizers

func (l *lexer) errorf(format string, args ...any) stateFn {
	l.emitError(format, args...)
	return nil
}

// state functions and related data

var keywords = map[string]itemType{
	"var":   K_VAR,
	"true":  K_TRUE,
	"false": K_FALSE,
	"not":   K_NOT,
}

const (
	digits      = "0123456789"
	lineComment = '#'
)

func lexStart(l *lexer) stateFn {

	switch r := l.next(); {
	case r == eof:
		l.emit(EOF)
		// ^^is it needed as a terminator in the yparser?
		// if not then l.ignore() it
		return nil
	case isShortToken(r):
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
		return lexInt
	default:
		return l.errorf("unknown char %#U", r)
	}
}

func lexSpace(l *lexer) stateFn {
	l.acceptRun(isSpace)
	l.ignore()
	return lexStart
}

func lexLineComment(l *lexer) stateFn {
	l.pos++
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
				l.emit(IDENT)
			}
			break loop
		}
	}
	return lexStart
}

func lexInt(l *lexer) stateFn {
	l.acceptRunFromSet(digits)
	l.emit(INT)
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
			if r == '\n' {
				l.line--
			}
			return l.errorf("unterminated quoted string")
		case '"':
			break loop
		}
	}
	l.emit(STR)
	return lexStart
}
