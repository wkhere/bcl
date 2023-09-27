package bcl

import (
	"fmt"
	"strings"
)

// external API
func parse(input string) (_ nTop, err error) {
	l := newLexer(input)
	p := yyNewParser()
	res := p.Parse(l)
	if res == 0 && l.err == nil {
		return p.(*yyParserImpl).lval.top, nil
	}
	if l.err != nil {
		return nTop{}, l.err
	}
	return nTop{}, fmt.Errorf("unknown error, parse state=%d\n", res)
}

// yacc->lex API

func (l *lexer) Lex(lval *yySymType) int {
	item := l.nextItem()
	//fmt.Printf("%v ", item) //dbg
	*lval = yySymType{t: strpos{s: item.val, pos: pos(item.pos)}}
	return int(item.typ)
}

func (l *lexer) Error(msg string) {
	if l.lastItem.err == nil {
		// error from yacc
		approx := l.lastItem
		if approx.val == "" {
			approx = l.prevItem
		}
		l.err = &errNearItem{approx, lineCalc(l.input), msg}
	} else {
		// error from lex
		l.err = &errAtItem{l.lastItem, lineCalc(l.input), msg}
	}
}

type lineCalc string

func (lc lineCalc) lineAt(pos int) int {
	return strings.Count(string(lc)[:pos], "\n") + 1
}

type errNearItem struct {
	item
	lineCalc
	msg string
}

func (e *errNearItem) Error() string {
	line := e.lineCalc.lineAt(e.item.pos)
	return fmt.Sprintf("line %d: %s near %q", line, e.msg, e.item.val)
}

type errAtItem struct {
	item
	lineCalc
	msg string
}

func (e *errAtItem) Error() string {
	line := e.lineCalc.lineAt(e.item.pos)
	return fmt.Sprintf("line %d: %s: %s", line, e.msg, e.item.err)
}
