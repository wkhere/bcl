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

var lexTab = []struct {
	input string
	items ii
}{
	{"", ii{{tEOF, "", nil, 1}}},
	{"!", ii{{tERR, "", fmt.Errorf("unknown char %#U", '!'), 1}}},
	{`"`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{"\"\n", ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{`"\`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
	{`"\a`, ii{{tERR, "", fmt.Errorf("unterminated quoted string"), 1}}},
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
