package bcl

import (
	"reflect"
	"strings"
	"testing"
)

type parsetc struct {
	input string
	top   nTop
	errs  string
}
type vmap = map[ident]expr

func pvalid(inp string, top nTop) parsetc { return parsetc{inp, top, ""} }
func perror(inp string, e string) parsetc { return parsetc{inp, nTop{}, e} }

func top1var(name ident, expr expr) nTop { return nTop{vars: vmap{name: expr}} }

var parseTab = []parsetc{
	pvalid(``, nTop{vars: vmap{}}),

	perror(`!`, `line 1: syntax error: unknown char`),
	perror(`foo `, `line 1: syntax error near "foo"`),
	perror(`foo bar`, `line 1: syntax error near "bar"`),

	pvalid(`var a = 1`, top1var("a", nIntLit(1))),
	pvalid(`var a = 1 + 2`, top1var("a", nBinOp{"+", nIntLit(1), nIntLit(2)})),

	perror(`+ 1`, `line 1: syntax error near "+"`),
	perror(`a + 1`, `line 1: syntax error near "+"`),
	perror(`var a + 1`, `line 1: syntax error near "+"`),

	perror(`var a=1aaa`, `line 1: syntax error near "aaa"`),

	pvalid(`var a=0xfce2`, top1var("a", nIntLit(64738))),
	pvalid(`var a=0xFCE2`, top1var("a", nIntLit(64738))),
	pvalid(`var a=0XFCE2`, top1var("a", nIntLit(64738))),
}

func TestParse(t *testing.T) {
	for i, tc := range parseTab {
		top, err := parse(tc.input)

		switch {
		case err != nil && tc.errs == "":
			t.Errorf("tc#%d unexpected error: %v", i, err)

		case err != nil && tc.errs != "":
			if !strings.Contains(err.Error(), tc.errs) {
				t.Errorf("tc#%d error mismatch\nhave %v\nwant %s",
					i, err, tc.errs,
				)
			}

		case err == nil && tc.errs != "":
			t.Errorf("tc#%d have no error, want error pattern %q", i, tc.errs)

		default:
			if !reflect.DeepEqual(top, tc.top) {
				t.Errorf("tc#%d mismatch:\nhave %+v\nwant %+v", i, top, tc.top)
			}
		}
	}
}
