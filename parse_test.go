package bcl

import (
	"reflect"
	"regexp"
	"strconv"
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
	// 1
	perror(`@`, `line 1: syntax error: unknown char`),
	perror(`!`, `line 1: syntax error: expected char '!' to start token "!="`),
	perror(`foo `, `line 1: syntax error near "foo"`),
	perror(`foo bar`, `line 1: syntax error near "bar"`),
	pvalid(`var a = 1`, top1var("a", nIntLit{1, 9})),
	// 5
	pvalid(`var a = 1 + 2`,
		top1var("a", nBinOp{"+", nIntLit{1, 9}, nIntLit{2, 13}}),
	),
	perror(`+ 1`, `line 1: syntax error near "+"`),
	perror(`a + 1`, `line 1: syntax error near "+"`),
	perror(`var a + 1`, `line 1: syntax error near "+"`),
	perror(`var a=1aaa`, `line 1: syntax error near "aaa"`),
	// 10
	pvalid(`var a=0x0`, top1var("a", nIntLit{0, 9})),
	pvalid(`var a=0xfce2`, top1var("a", nIntLit{64738, 12})),
	pvalid(`var a=0xFCE2`, top1var("a", nIntLit{64738, 12})),
	pvalid(`var a=0XFCE2`, top1var("a", nIntLit{64738, 12})),
	pvalid(`var a=0xFC00 + 0xE2`,
		top1var("a", nBinOp{"+", nIntLit{64512, 12}, nIntLit{226, 19}}),
	),
	// 15
	pvalid(`var a= 1==2`,
		top1var("a", nBinOp{"==", nIntLit{1, 8}, nIntLit{2, 11}}),
	),
	pvalid(`var a= 1!=2`,
		top1var("a", nBinOp{"!=", nIntLit{1, 8}, nIntLit{2, 11}}),
	),
	pvalid(`var a= 1==2==false`, top1var("a",
		nBinOp{"==",
			nBinOp{"==", nIntLit{1, 8}, nIntLit{2, 11}},
			nBoolLit{false, 18},
		},
	)),
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

func TestParseMultipleLines(t *testing.T) {
	const (
		input1 = `
		var x = 1
		testBlock "foo" []`

		input2 = `
		this is wr&%#@%&6ong
		`

		input3 = `
		= abc def ghi
		`

		input4 = `
		!%^$@%!
		`
	)

	var linep = regexp.MustCompile(`^line (\d+):`)

	var tab = []struct {
		input    string
		failLine int
	}{
		{input1, 3},
		{input2, 2},
		{input3, 2},
		{input4, 2},
	}

	for i, tc := range tab {
		_, err := parse(tc.input)
		if err == nil {
			t.Errorf("tc#%d no error but expected one", i)
			continue
		}

		m := linep.FindStringSubmatch(err.Error())
		if len(m) == 0 {
			t.Errorf("tc#%d expected to find line numer in: %v", i, err)
			continue
		}
		line, err2 := strconv.Atoi(m[1])
		if err2 != nil {
			t.Errorf("tc#%d can't parse line number in: %v\n%v", i, err, err2)
			continue
		}
		if line != tc.failLine {
			t.Errorf(
				"tc#%d mismatch: have line %d with error, want %d; error:\n%v",
				i, line, tc.failLine, err,
			)
		}
	}
}
