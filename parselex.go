package bcl

import "fmt"

// external API
func parse(input []byte) (_ nTop, err error) {
	l := newLexer(string(input))
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
	*lval = yySymType{s: item.val}
	return int(item.typ)
}

func (l *lexer) Error(e string) {
	if l.lastItem.err == nil {
		// error from yacc
		approx := l.lastItem.val
		if approx == "" {
			approx = l.prevItem.val
		}
		l.err = fmt.Errorf("line %d: %s near %q", l.line, e, approx)
	} else {
		// error from lex
		l.err = fmt.Errorf("line %d: %s: %s", l.line, e, l.lastItem.err)
	}
}
