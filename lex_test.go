package bcl

import (
	"fmt"
	"reflect"
	"testing"
)

type istream <-chan item

func (s istream) collect() (a []item) {
	for x := range s {
		a = append(a, x)
	}
	return a
}

// shorter syntax in tab literals:
type ii = []item

func teof(pos int) item { return item{tEOF, "", nil, pos} }
func terrchar(c rune, line int) item {
	return item{tERR, "", fmt.Errorf("unknown char %#U", c), line}
}

var lexTab = []struct {
	input string
	items ii
}{
	{"", ii{teof(0)}},

	{"@", ii{terrchar('@', 1)}},
	{`"`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 2}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 2}}},
	{`"\`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 2}}},
	{`"\a`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 3}}},
	{`1234`, ii{{tINT, "1234", nil, 4}, teof(4)}},
	{`12.34`, ii{{tFLOAT, "12.34", nil, 5}, teof(5)}},
	{`1234e10`, ii{{tFLOAT, "1234e10", nil, 7}, teof(7)}},
	{`1234E10`, ii{{tFLOAT, "1234E10", nil, 7}, teof(7)}},

	{`1234e+10`, ii{{tFLOAT, "1234e+10", nil, 8}, teof(8)}},
	{`1234e-10`, ii{{tFLOAT, "1234e-10", nil, 8}, teof(8)}},
	{`12.34e10`, ii{{tFLOAT, "12.34e10", nil, 8}, teof(8)}},
	{`12.34e+10`, ii{{tFLOAT, "12.34e+10", nil, 9}, teof(9)}},
	{`12.34e-10`, ii{{tFLOAT, "12.34e-10", nil, 9}, teof(9)}},
	{`12.`, ii{{tERR, "", fmt.Errorf("need more digits after a dot"), 3}}},
	{`12e`, ii{{tERR, "", fmt.Errorf("need more digits for an exponent"), 3}}},
	{`0x10`, ii{{tINT, "0x10", nil, 4}, teof(4)}},
	{`0X10`, ii{{tINT, "0X10", nil, 4}, teof(4)}},
	{`0x10.0`, ii{{tINT, "0x10", nil, 4}, terrchar('.', 5)}},

	{`>`, ii{{'>', ">", nil, 1}, teof(1)}},
	{`>=`, ii{{tGE, ">=", nil, 2}, teof(2)}},
	{`< 5`, ii{{'<', "<", nil, 1}, {tINT, "5", nil, 3}, teof(3)}},
	{`<= 5`, ii{{tLE, "<=", nil, 2}, {tINT, "5", nil, 4}, teof(4)}},
	{`!<`, ii{
		{tERR, "", fmt.Errorf(`expected char '!' to start token "!="`), 1},
	}},
	{`{`, ii{{'{', "{", nil, 1}, teof(1)}},
}

func TestLexer(t *testing.T) {
	for i, tc := range lexTab {
		l := newLexer(tc.input)
		items := istream(l.items).collect()
		if !reflect.DeepEqual(items, tc.items) {
			t.Errorf("tc#%d mismatch:\nhave %v\nwant %v", i, items, tc.items)
		}
	}
}
