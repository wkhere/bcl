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

func teof(line int) item { return item{tEOF, "", nil, line} }
func terrchar(c rune, line int) item {
	return item{tERR, "", fmt.Errorf("unknown char %#U", c), line}
}

var lexTab = []struct {
	input string
	items ii
}{
	{"", ii{teof(1)}},

	{"!", ii{terrchar('!', 1)}},

	{`"`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{`"\`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{`"\a`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},

	{`1234`, ii{{tINT, "1234", nil, 1}, teof(1)}},
	{`12.34`, ii{{tFLOAT, "12.34", nil, 1}, teof(1)}},
	{`1234e10`, ii{{tFLOAT, "1234e10", nil, 1}, teof(1)}},
	{`1234E10`, ii{{tFLOAT, "1234E10", nil, 1}, teof(1)}},
	{`1234e+10`, ii{{tFLOAT, "1234e+10", nil, 1}, teof(1)}},
	{`1234e-10`, ii{{tFLOAT, "1234e-10", nil, 1}, teof(1)}},
	{`12.34e10`, ii{{tFLOAT, "12.34e10", nil, 1}, teof(1)}},
	{`12.34e+10`, ii{{tFLOAT, "12.34e+10", nil, 1}, teof(1)}},
	{`12.34e-10`, ii{{tFLOAT, "12.34e-10", nil, 1}, teof(1)}},
	{`12.`, ii{{tERR, "", fmt.Errorf("need more digits after a dot"), 1}}},
	{`12e`, ii{{tERR, "", fmt.Errorf("need more digits for an exponent"), 1}}},

	{`0x10`, ii{{tINT, "0x10", nil, 1}, teof(1)}},
	{`0x10.0`, ii{{tINT, "0x10", nil, 1}, terrchar('.', 1)}},
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
