package bcl

import (
	"reflect"
	"strings"
	"testing"
)

type evaltc struct {
	input    string
	inputRaw *nTop
	blks     []Block
	errs     string
}

// shorter syntax in literals:
type emap = map[string]any
type bb = []Block

func evalid(in string, bb []Block) evaltc { return evaltc{in, nil, bb, ""} }
func eerror(in string, e string) evaltc   { return evaltc{in, nil, bb{}, e} }

func evalidRaw(in nTop, bb []Block) evaltc { return evaltc{"", &in, bb, ""} }
func eerrorRaw(in nTop, e string) evaltc   { return evaltc{"", &in, bb{}, e} }

var evalTab = []evaltc{
	evalid(``, bb{}),

	eerror(`var y=x`, "var x not defined"),
	eerror(`var x=x`, "var x: cycle"),
	eerror(`blk "foo" {a=x}`, "var x not defined"),

	evalid(`var a=1+1  blk "b" {a=a}`, bb{{"blk", "b", emap{"a": 2}}}),
	evalid(`var a=1*(3-5)/-2  blk "b" {a=a}`, bb{{"blk", "b", emap{"a": 1}}}),

	eerror(`var s="a"-2`, `op "-": invalid types: string, int`),
	eerror(`var s=2-"a"`, `op "-": invalid types: int, string`),

	evalid(`var s="a"+2  blk "b" {s=s}`, bb{{"blk", "b", emap{"s": "a2"}}}),
	eerror(`var s=2+"a"`, `op "+": invalid types: int, string`),

	evalid(`var s="a"*2  blk "b" {s=s}`, bb{{"blk", "b", emap{"s": "aa"}}}),
	eerror(`var s=2*"a"`, `op "*": invalid types: int, string`),

	eerror(`var s="a"/2`, `op "/": invalid types: string, int`),
	eerror(`var s=2/"a"`, `op "/": invalid types: int, string`),

	evalid(`var x=-1  blk "b" {x=x}`, bb{{"blk", "b", emap{"x": -1}}}),
	eerror(`var x=-"a"`, `op "-": invalid type: string`),

	evalid(`var x=not false  blk "b" {x=x}`, bb{{"blk", "b", emap{"x": true}}}),
	eerror(`var x=not 1`, `op "not": invalid type: int`),
	eerror(`var x=not "foo"`, `op "not": invalid type: string`),

	eerror(`var x=-(not 1)`, `op "not": invalid type: int`),
	eerror(`var x=not(1-false)`, `op "-": invalid types: int, bool`),

	eerror(`var x=1+(not 1)`, `op "not": invalid type: int`),
	eerror(`var x=(not 1)+1`, `op "not": invalid type: int`),

	eerrorRaw(
		nTop{vars: vmap{"a": nUnOp{"@", nIntLit(1)}}},
		`unknown op "unary @"`,
	),
	eerrorRaw(
		nTop{vars: vmap{"a": nBinOp{"@", nIntLit(1), nIntLit(2)}}},
		`unknown op "binary @"`,
	),

	eerror(`var x=1/0`, `division by zero`),
}

func TestEval(t *testing.T) {
	for i, tc := range evalTab {
		var blks []Block
		var err error

		if tc.inputRaw != nil {
			blks, err = eval(tc.inputRaw)
		} else {
			blks, err = Interpret([]byte(tc.input))
		}

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
			if !reflect.DeepEqual(blks, tc.blks) {
				t.Errorf("tc#%d mismatch:\nhave %+v\nwant %+v", i, blks, tc.blks)
			}
		}
	}
}
