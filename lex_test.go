package bcl

import (
	"fmt"
	"reflect"
	"testing"
)

type tstream <-chan token

func (s tstream) collect() (a []token) {
	for x := range s {
		a = append(a, x)
	}
	return a
}

// shorter syntax in tab literals:
type tt = []token

func teof(pos int) token  { return token{tEOF, "", nil, pos} }
func tfail(pos int) token { return token{tFAIL, "", nil, pos} }
func terrchar(c rune, pos int) token {
	return token{tERR, "", fmt.Errorf("unknown char %#U", c), pos}
}
func terrinvalid(s string, pos int) token {
	return token{tERR, "", fmt.Errorf("invalid syntax `%s`", s), pos}
}

var errUnterminatedQuote = fmt.Errorf("unterminated quoted string")

var lexTab = []struct {
	i      int
	input  string
	tokens tt
}{
	{0, "", tt{teof(0)}},

	{1, "@", tt{terrchar('@', 1), tfail(1)}},
	{2, `"`, tt{{tERR, "", errUnterminatedQuote, 1}, tfail(1)}},
	{3, "\"\n", tt{{tERR, "", errUnterminatedQuote, 2}, tfail(2)}},
	{4, "\"\n", tt{{tERR, "", errUnterminatedQuote, 2}, tfail(2)}},
	{5, `"\`, tt{{tERR, "", errUnterminatedQuote, 2}, tfail(2)}},
	{6, `"\a`, tt{{tERR, "", errUnterminatedQuote, 3}, tfail(3)}},

	{7, `1234`, tt{{tINT, "1234", nil, 4}, teof(4)}},
	{8, `12.34`, tt{{tFLOAT, "12.34", nil, 5}, teof(5)}},
	{9, `1234e10`, tt{{tFLOAT, "1234e10", nil, 7}, teof(7)}},
	{10, `1234E10`, tt{{tFLOAT, "1234E10", nil, 7}, teof(7)}},
	{11, `1234e+10`, tt{{tFLOAT, "1234e+10", nil, 8}, teof(8)}},
	{12, `1234e-10`, tt{{tFLOAT, "1234e-10", nil, 8}, teof(8)}},
	{13, `12.34e10`, tt{{tFLOAT, "12.34e10", nil, 8}, teof(8)}},
	{14, `12.34e+10`, tt{{tFLOAT, "12.34e+10", nil, 9}, teof(9)}},
	{15, `12.34e-10`, tt{{tFLOAT, "12.34e-10", nil, 9}, teof(9)}},
	{16, `12.`, tt{{tERR, "", fmt.Errorf("need more digits after a dot"), 3}, tfail(3)}},
	{17, `12e`, tt{{tERR, "", fmt.Errorf("need more digits for an exponent"), 3}, tfail(3)}},

	{18, `0x10`, tt{{tINT, "0x10", nil, 4}, teof(4)}},
	{19, `0X10`, tt{{tINT, "0X10", nil, 4}, teof(4)}},
	{20, `0x10.0`, tt{terrinvalid("0x10.", 5), tfail(5)}},

	{21, `>`, tt{{tGT, ">", nil, 1}, teof(1)}},
	{22, `>=`, tt{{tGE, ">=", nil, 2}, teof(2)}},
	{23, `< 5`, tt{{tLT, "<", nil, 1}, {tINT, "5", nil, 3}, teof(3)}},
	{24, `<= 5`, tt{{tLE, "<=", nil, 2}, {tINT, "5", nil, 4}, teof(4)}},
	{25, `!<`, tt{
		{tERR, "", fmt.Errorf(`expected char '!' to start token "!="`), 1},
		tfail(1),
	}},

	{26, `{}`, tt{{tLCURLY, "{", nil, 1}, {tRCURLY, "}", nil, 2}, teof(2)}},
	{27, `()`, tt{{tLPAREN, "(", nil, 1}, {tRPAREN, ")", nil, 2}, teof(2)}},
	//{28, `[]`, tt{{tLBRACKET, "[", nil, 1}, {tRBRACKET, "]", nil, 2}, teof(2)}},

	{29, `or`, tt{{tOR, "or", nil, 2}, teof(2)}},
	{30, `and`, tt{{tAND, "and", nil, 3}, teof(3)}},
	{31, `"foo"`, tt{{tSTR, `"foo"`, nil, 5}, teof(5)}},

	{32, "#a", tt{teof(2)}},
	{33, "#a\n", tt{teof(3)}},
	{34, "#", tt{teof(1)}},
	{35, "#\n", tt{teof(2)}},

	{36, "a", tt{{tIDENT, "a", nil, 1}, teof(1)}},
	{37, "a\n", tt{{tIDENT, "a", nil, 1}, teof(2)}},
	{38, "a\nb", tt{{tIDENT, "a", nil, 1}, {tIDENT, "b", nil, 3}, teof(3)}},
	{39, "a\nb\n", tt{{tIDENT, "a", nil, 1}, {tIDENT, "b", nil, 3}, teof(4)}},
	{40, "a\nbb\nc", tt{
		{tIDENT, "a", nil, 1}, {tIDENT, "bb", nil, 4}, {tIDENT, "c", nil, 6},
		teof(6),
	}},
	{41, "a\nbb\nc\n\n", tt{
		{tIDENT, "a", nil, 1}, {tIDENT, "bb", nil, 4}, {tIDENT, "c", nil, 6},
		teof(8),
	}},
	{42, "a\nbb\nc\n\ndd", tt{
		{tIDENT, "a", nil, 1}, {tIDENT, "bb", nil, 4}, {tIDENT, "c", nil, 6},
		{tIDENT, "dd", nil, 10}, teof(10),
	}},
	{43, "foobar quux1 finito", tt{
		{tIDENT, "foobar", nil, 6}, {tIDENT, "quux1", nil, 12},
		{tIDENT, "finito", nil, 19}, teof(19),
	}},

	{44, "a42", tt{{tIDENT, "a42", nil, 3}, teof(3)}},
	{45, `"foo"`, tt{{tSTR, `"foo"`, nil, 5}, teof(5)}},

	{46, "42q", tt{terrinvalid("42q", 3), tfail(3)}},
	{47, `42"q"`, tt{terrinvalid(`42"`, 3), tfail(3)}},
	{48, "42.0q", tt{terrinvalid("42.0q", 5), tfail(5)}},
	{49, "0x42q", tt{terrinvalid("0x42q", 5), tfail(5)}},
	{50, `42.0"q"`, tt{terrinvalid(`42.0"`, 5), tfail(5)}},
	{51, `0x42"q"`, tt{terrinvalid(`0x42"`, 5), tfail(5)}},
	{52, `var"q"`, tt{terrinvalid(`var"`, 4), tfail(4)}},
	{53, `foo"q"`, tt{terrinvalid(`foo"`, 4), tfail(4)}},
	{54, `"foo"1`, tt{terrinvalid(`"foo"1`, 6), tfail(6)}},
	{55, `"foo"q`, tt{terrinvalid(`"foo"q`, 6), tfail(6)}},
}

func TestLexerSingleInput(t *testing.T) {
	for _, tc := range lexTab {
		c := make(chan string, 1)
		c <- tc.input
		close(c)
		l := newLexer(c, dummyLcUpd)
		res := tstream(l.tokens).collect()
		if !reflect.DeepEqual(res, tc.tokens) {
			t.Errorf("tc#%d mismatch:\nhave %v\nwant %v", tc.i, res, tc.tokens)
		}
	}
}

func TestLexerManyInputs(t *testing.T) {
	for _, tc := range lexTab {
		chunks := eachN(tc.input, 4)
		c := make(chan string, len(chunks))
		for _, s := range chunks {
			c <- s
		}
		close(c)
		l := newLexer(c, dummyLcUpd)
		res := tstream(l.tokens).collect()
		if !reflect.DeepEqual(res, tc.tokens) {
			t.Errorf("tc#%d mismatch:\nhave %v\nwant %v", tc.i, res, tc.tokens)
		}
	}
}

func ExampleLexer() {
	runExample("0")
	runExample("1")
	runExample("-3.14")
	runExample("@")
	// Output:
	// {tINT "0" 1}{tEOF "" 1}
	// {tINT "1" 1}{tEOF "" 1}
	// {tMINUS "-" 1}{tFLOAT "3.14" 5}{tEOF "" 5}
	// {tERR "unknown char U+0040 '@'" 1}{tFAIL "" 1}
}

func runExample(s string) {
	c := make(chan string, 1)
	c <- s
	close(c)
	l := newLexer(c, dummyLcUpd)
	for r := range l.tokens {
		fmt.Print(r)
	}
	fmt.Println()

}

func dummyLcUpd(string, int) {}

func eachN(s string, n int) (chunks []string) {
	chunks = make([]string, 0, len(s)/n+1)
	for len(s) > 0 {
		j := min(len(s), n)
		chunks = append(chunks, s[:j])
		s = s[j:]
	}
	return
}
