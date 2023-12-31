package bcl

import (
	"reflect"
	"strings"
	"testing"
)

type evaltc struct {
	input string
	ast   *nTop
	blks  []Block
	errs  string
}

// shorter syntax in literals:
type emap = map[string]any
type bb = []Block

func evalid(in string, bb []Block) evaltc { return evaltc{in, nil, bb, ""} }
func eerror(in string, e string) evaltc   { return evaltc{in, nil, bb{}, e} }

func evalidAst(in nTop, bb []Block) evaltc { return evaltc{"", &in, bb, ""} }
func eerrorAst(in nTop, e string) evaltc   { return evaltc{"", &in, bb{}, e} }

var (
	xtrue  = bb{{"blk", "b", emap{"x": true}}}
	xfalse = bb{{"blk", "b", emap{"x": false}}}
)
var evalTab = []evaltc{
	evalid(``, bb{}),
	// 1
	eerror(`var y=x`, "var x not defined"),
	eerror(`var x=x`, "var x: cycle"),
	eerror(`def blk "foo" {a=x}`, "var x not defined"),
	evalid(`var a=1+1  def blk "b" {a=a}`, bb{{"blk", "b", emap{"a": 2}}}),
	evalid(`var a=1*(3-5)/-2  def blk "b" {a=a}`, bb{{"blk", "b", emap{"a": 1}}}),
	// 5
	eerror(`var s="a"-2`, `op "-": invalid types: string, int`),
	eerror(`var s=2-"a"`, `op "-": invalid types: int, string`),
	evalid(`var s="a"+2  def blk "b" {s=s}`, bb{{"blk", "b", emap{"s": "a2"}}}),
	eerror(`var s=2+"a"`, `op "+": invalid types: int, string`),
	evalid(`var s="a"*2  def blk "b" {s=s}`, bb{{"blk", "b", emap{"s": "aa"}}}),
	// 10
	eerror(`var s=2*"a"`, `op "*": invalid types: int, string`),
	eerror(`var s="a"/2`, `op "/": invalid types: string, int`),
	eerror(`var s=2/"a"`, `op "/": invalid types: int, string`),
	evalid(`var x=-1   def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": -1}}}),
	evalid(`var x=--1  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 1}}}),
	// 15
	eerror(`var x=-"a"`, `op "-": invalid type: string`),
	evalid(`var x=not false  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": true}}}),
	eerror(`var x=not 1`, `op "not": invalid type: int`),
	eerror(`var x=not "foo"`, `op "not": invalid type: string`),
	eerror(`var x=-(not 1)`, `op "not": invalid type: int`),
	// 20
	eerror(`var x=not(1-false)`, `op "-": invalid types: int, bool`),
	eerror(`var x=1+(not 1)`, `op "not": invalid type: int`),
	eerror(`var x=(not 1)+1`, `op "not": invalid type: int`),
	evalid(`var x=1.23    def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +1.23}}}),
	evalid(`var x=--1.23  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +1.23}}}),
	// 25
	evalid(`var x=-1.23   def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": -1.23}}}),
	evalid(`var x=1+1.23  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +2.23}}}),
	evalid(`var x=1.23+1  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +2.23}}}),
	evalid(`var x=1.2+1.5 def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +2.7}}}),
	eerror(`var s="a"+2.0  `, `invalid types: string, float64`),
	// 30
	evalid(`var x=1.2-10  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": -8.8}}}),
	evalid(`var x=10-1.2  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +8.8}}}),
	evalid(`var x=2.5-0.5 def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": +2.0}}}),
	eerror(`var s="a"-2.0  `, `invalid types: string, float64`),
	evalid(`var x=10*1.2  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 12.0}}}),
	// 35
	evalid(`var x=1.2*10   def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 12.0}}}),
	evalid(`var x=1.2*0.5  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 0.6}}}),
	eerror(`var s="a"*2.0  `, `invalid types: string, float64`),
	evalid(`var x=10/2.5   def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 4.0}}}),
	evalid(`var x=1.2/10   def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 0.12}}}),
	// 40
	evalid(`var x=1.2/0.5  def blk "b" {x=x}`, bb{{"blk", "b", emap{"x": 2.4}}}),
	eerror(`var x=true/2.0 `, `invalid types: bool, float64`),
	eerror(`var x=2.0/true `, `invalid types: float64, bool`),
	eerror(`var s="a"/2.0  `, `invalid types: string, float64`),
	eerror(`var s=2.0/"a"  `, `invalid types: float64, string`),
	// 45
	eerror(`var x=0/0    `, `division by zero`),
	eerror(`var x=1/0    `, `division by zero`),
	eerror(`var x=1.0/0  `, `division by zero`),
	eerror(`var x=1/0.0  `, `division by zero`),
	eerror(`var x=1.0/0.0`, `division by zero`),
	// 50
	evalid(`var x= 1==1   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1==2   def blk "b" {x=x}`, xfalse),
	evalid(`var x= 1!=1   def blk "b" {x=x}`, xfalse),
	evalid(`var x= 1!=2   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1==1.0 def blk "b" {x=x}`, xtrue),
	// 55
	evalid(`var x= 1.0==1    def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0!=1    def blk "b" {x=x}`, xfalse),
	evalid(`var x= 1!=1.0    def blk "b" {x=x}`, xfalse),
	evalid(`var x= 1.0==1.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0!=1.0  def blk "b" {x=x}`, xfalse),
	// 60
	eerror(`var x= "s"==1`, `invalid types: string, int`),
	eerror(`var x= "s"!=1`, `invalid types: string, int`),
	eerror(`var x= 1=="s"`, `invalid types: int, string`),
	eerror(`var x= 1!="s"`, `invalid types: int, string`),
	eerror(`var x= "s"!=1`, `invalid types: string, int`),
	// 65
	evalid(`var x= ""==""    def blk "b" {x=x}`, xtrue),
	evalid(`var x= "a"=="a"  def blk "b" {x=x}`, xtrue),
	evalid(`var x= "a"=="b"  def blk "b" {x=x}`, xfalse),
	evalid(`var x= "a"==""   def blk "b" {x=x}`, xfalse),
	evalid(`var x= ""=="a"   def blk "b" {x=x}`, xfalse),
	// 70
	evalid(`var x= 1.0!=1.0  def blk "b" {x=x}`, xfalse),
	evalid(`var x= 1.0!=1.2  def blk "b" {x=x}`, xtrue),
	evalid(`var x= ""!=""    def blk "b" {x=x}`, xfalse),
	evalid(`var x= "a"!="a"  def blk "b" {x=x}`, xfalse),
	evalid(`var x= "a"!="b"  def blk "b" {x=x}`, xtrue),
	// 75
	evalid(`var x= true==true     def blk "b" {x=x}`, xtrue),
	evalid(`var x= true==false    def blk "b" {x=x}`, xfalse),
	evalid(`var x= true==(1==1)   def blk "b" {x=x}`, xtrue),
	evalid(`var x= (1==1)==true   def blk "b" {x=x}`, xtrue),
	evalid(`var x= (1==1)==(2==2) def blk "b" {x=x}`, xtrue),
	// 80
	evalid(`var x= true!=false    def blk "b" {x=x}`, xtrue),
	evalid(`var x= true!=true     def blk "b" {x=x}`, xfalse),
	evalid(`var x= true!=(1!=1)   def blk "b" {x=x}`, xtrue),
	evalid(`var x= (1==1)!=(2==2) def blk "b" {x=x}`, xfalse),
	evalid(`var x= (1==1)!=(2!=2) def blk "b" {x=x}`, xtrue),
	// 85
	evalid(`var x= 1<2   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1>2   def blk "b" {x=x}`, xfalse),
	evalid(`var x= 2>1   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1<=1  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1<=2  def blk "b" {x=x}`, xtrue),
	// 90
	evalid(`var x= 1>=1   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 2>=1   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1<2.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0<2  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 2>1.0  def blk "b" {x=x}`, xtrue),
	// 95
	evalid(`var x= 2.0>1    def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0<2.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 2.0>1.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1<=1.0   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1<=2.0   def blk "b" {x=x}`, xtrue),
	// 100
	evalid(`var x= 1.0<=1   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0<=2   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1>=1.0   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 2>=1.0   def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0>=1   def blk "b" {x=x}`, xtrue),
	// 105
	evalid(`var x= 2.0>=1    def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0<=1.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0<=2.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 1.0>=1.0  def blk "b" {x=x}`, xtrue),
	evalid(`var x= 2.0>=1.0  def blk "b" {x=x}`, xtrue),
	// 110
	evalid(`var x= "a"<"b"   def blk "b" {x=x}`, xtrue),
	evalid(`var x= "b">"a"   def blk "b" {x=x}`, xtrue),
	evalid(`var x= "a"<="a"  def blk "b" {x=x}`, xtrue),
	evalid(`var x= "a"<="b"  def blk "b" {x=x}`, xtrue),
	evalid(`var x= "b">="b"  def blk "b" {x=x}`, xtrue),
	// 115
	evalid(`var x= "b">="a"  def blk "b" {x=x}`, xtrue),
	eerror(`var x= false<true`, `invalid types: bool, bool`),
	eerror(`var x= false>true`, `invalid types: bool, bool`),
	eerror(`var x= false<=true`, `invalid types: bool, bool`),
	eerror(`var x= false>=true`, `invalid types: bool, bool`),
	// 120
	eerror(`var x= true<1`, `invalid types: bool, int`),
	eerror(`var x= true>1`, `invalid types: bool, int`),
	eerror(`var x= true<=1`, `invalid types: bool, int`),
	eerror(`var x= true>=1`, `invalid types: bool, int`),
	eerror(`var x= 1<true`, `invalid types: int, bool`),
	// 125
	eerror(`var x= 1>true`, `invalid types: int, bool`),
	eerror(`var x= 1<=true`, `invalid types: int, bool`),
	eerror(`var x= 1>=true`, `invalid types: int, bool`),
	eerror(`var x= 1<"a"`, `invalid types: int, string`),
	eerror(`var x= 1>"a"`, `invalid types: int, string`),
	// 130
	eerror(`var x= 1<="a"`, `invalid types: int, string`),
	eerror(`var x= 1>="a"`, `invalid types: int, string`),
	eerror(`var x= "a"<1`, `invalid types: string, int`),
	eerror(`var x= "a">1`, `invalid types: string, int`),
	eerror(`var x= "a"<=1`, `invalid types: string, int`),
	// 135
	eerror(`var x= "a">=1`, `invalid types: string, int`),
	eerror(`var x= "a"<1.0`, `invalid types: string, float`),
	eerror(`var x= "a">1.0`, `invalid types: string, float`),
	eerror(`var x= 1.0<"a"`, `invalid types: float64, string`),
	eerror(`var x= 1.0>"a"`, `invalid types: float64, string`),
	// 140
	evalid(`var x= true or true    def blk "b" {x=x}`, xtrue),
	evalid(`var x= true or false   def blk "b" {x=x}`, xtrue),
	evalid(`var x= false or true   def blk "b" {x=x}`, xtrue),
	evalid(`var x= false or false  def blk "b" {x=x}`, xfalse),
	evalid(`var x= false or false or true  def blk "b" {x=x}`, xtrue),
	// 145
	evalid(`var x= true and true    def blk "b" {x=x}`, xtrue),
	evalid(`var x= true and false   def blk "b" {x=x}`, xfalse),
	evalid(`var x= false and true   def blk "b" {x=x}`, xfalse),
	evalid(`var x= false and false  def blk "b" {x=x}`, xfalse),
	evalid(`var x= true and true and false  def blk "b" {x=x}`, xfalse),
	// 150
	evalid(`var x= 1==1 or 1==2 and 1==3    def blk "b" {x=x}`, xtrue),
	evalid(`var x= (1==1 or 1==2) and 1==3  def blk "b" {x=x}`, xfalse),
	evalid(`var x= not false and true  def blk "b" {x=x}`, xtrue),
	evalid(`var x= not false or true   def blk "b" {x=x}`, xtrue),
	evalid(`var x= not (false or true) def blk "b" {x=x}`, xfalse),
	// 155
	eerror(`var x= 1 or false`, `invalid type of 1st operand: int`),
	eerror(`var x= false or 1`, `invalid type of 2nd operand: int`),
	eerror(`var x= 1 or false or 1`, `invalid type of 1st operand: int`),
	eerror(`var x= false or (false or 1)`, `invalid type of 2nd operand: int`),
	eerror(`var x= 1<"a" or false`, `invalid types: int, string`),

	eerrorAst(
		nTop{vars: vmap{"a": nUnOp{"@", nIntLit{1, 0}}}},
		`unknown op "unary @"`,
	),
	eerrorAst(
		nTop{vars: vmap{"a": nBinOp{"@", nIntLit{1, 0}, nIntLit{2, 0}}}},
		`unknown op "binary @"`,
	),
	eerrorAst(
		nTop{vars: vmap{"a": nSCOp{"@", nIntLit{1, 0}, nIntLit{2, 0}}}},
		`unknown op "short circuit @"`,
	),
}

func TestEval(t *testing.T) {
	for i, tc := range evalTab {
		var blks []Block
		var err error

		if tc.ast != nil {
			blks, err = eval(tc.ast, lineCalc(tc.input))
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

func TestEvalInvalidStage(t *testing.T) {
	env := env{}
	env.stage = -1

	_, err := nVarRef{"foo", 0}.eval(&env)

	if !strings.Contains(err.Error(), "invalid eval stage") {
		t.Errorf("unknown error: %s", err)
	}
}
