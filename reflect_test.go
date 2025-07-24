package bcl

import (
	"io"
	"reflect"
	"regexp"
	"testing"
)

type Struct1 struct {
	Name         string
	Field1       int
	Field2       string
	Status       bool `bcl:"field3"`
	AnotherField float64
	Other        int `foo:"bar" bcl:"other_field"`

	Inner struct {
		Name string
		Text string `bcl:"field4"`
	}

	Inner2 struct {
		Name string
		Text string `bcl:"field5"`
	} `bcl:"inner.foo"`
}

func simpleUnmarshal(dest any) error {
	const bcl = `
	def struct1 "foo" {
		field1 = 10
		field2 = "abc"
		field3 = true

		def inner {
			field4 = "inner-"+(field1+1)
		}

		def inner "foo" {
			field5 = "inner-"+(field1+2)
		}

		another_field = 10.2
		other_field = 42
	}

	bind struct1
	`
	return Unmarshal([]byte(bcl), dest)
}

func TestSimpleUnmarshal(t *testing.T) {
	var x Struct1

	err := simpleUnmarshal(&x)
	if err != nil {
		t.Error(err)
		return
	}
	if v := x.Field1; v != 10 {
		t.Errorf("expected Field1==10")
	}
	if v := x.Other; v != 42 {
		t.Errorf("expected Other==42")
	}
	if v := x.Inner.Text; v != "inner-11" {
		t.Errorf("expected Inner==%q", "inner-11")
	}
	if v := x.Inner2.Text; v != "inner-12" {
		t.Errorf("expected Inner2==%q", "inner-12")
	}
	t.Logf("%+v", x)
}

func BenchmarkSimpleUnmarshal(b *testing.B) {
	var x Struct1

	for i := 0; i < b.N; i++ {
		simpleUnmarshal(&x)
	}
}

type reflecttc struct {
	input string
	dest  any
	want  any
	errp  *regexp.Regexp
}

func rvalid(in string, d, w any) reflecttc {
	return reflecttc{in, d, w, nil}
}
func rerror(in string, d any, e string) reflecttc {
	return reflecttc{in, d, nil, regexp.MustCompile(e)}
}

type S struct {
	Name string
	X    int
}

type u struct{ S }       // umbrella for S
type um struct{ Ss []S } // umbrella for multiple S

type S2 struct {
	Name string
	X    int `bcl:"y"`
}

type S3 struct {
	X    int
	Name string // note the name is not the first field
}

type S4 struct {
	X, Y     int
	DontCare any
}

type S5 struct {
	X float64
	I int
}

var reflectTab = []reflecttc{

	rerror(``, nil, "no binding"),
	rerror(``, 1, "no binding"),
	rerror(``, struct{}{}, "no binding"),
	rerror(``, &struct{}{}, "no binding"),
	rerror(``, &[]struct{}{}, "no binding"),

	rerror(`bind any`, nil, "no blocks of type any"),
	rerror(`bind any`, 1, "no blocks of type any"),
	rerror(`bind any`, struct{}{}, "no blocks of type any"),
	rerror(`bind any`, &struct{}{}, "no blocks of type any"),
	rerror(`bind any`, &[]struct{}{}, "no blocks of type any"),

	// 10:
	rerror(`def any{}; bind any`, nil, "expected pointer"),
	rerror(`def any{}; bind any`, 1, "expected pointer, have: int"),
	rerror(`def any{}; bind any`, struct{}{}, "expected pointer, have: struct"),
	rvalid(`def any{}; bind any`, &struct{}{}, &struct{}{}),
	rerror(`def any{}; bind any`, &[]struct{}{}, "expected struct, have: slice"),

	// 15:
	rerror(`def any{}; bind any`, nil, "expected pointer"),
	rerror(`def any{}; bind any`, struct{}{}, "expected pointer, have: struct"),
	rerror(`def any{}; bind any:all`, &struct{}{}, "expected slice, have: struct"),
	rvalid(`def any{}; bind any:all`, &[]struct{}{}, &[]struct{}{{}}), // empty struct inside
	// the next one is nasty: would panic if there was no checking for a slice element:
	rerror(`def int{}; bind int:all`, &[]int{},
		"slice element deref: expected struct, have: int",
	),

	// 20:
	rvalid(`def any{x=1}; bind any`, &struct{ X int }{}, &struct{ X int }{X: 1}),
	rvalid(`def any{x=1}; bind any:all`, &[]struct{ X int }{}, &[]struct{ X int }{{X: 1}}),
	rvalid(`def any "foo" {}; bind any`,
		&struct{ Name string }{},
		&struct{ Name string }{Name: "foo"},
	),
	rvalid(`def any "foo" {}; bind any:all`,
		&[]struct{ Name string }{},
		&[]struct{ Name string }{{Name: "foo"}},
	),
	rerror(`def any "foo" {}; bind any`,
		&struct{}{},
		`field mapping for "Name" not found in struct`,
	),

	// 25:
	rvalid(`def any "foo" {s="quux"}; bind any`,
		&struct{ Name, S string }{},
		&struct{ Name, S string }{Name: "foo", S: "quux"},
	),
	rerror(`def any "foo" {x=10}; bind any`,
		&struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`def any "foo" {x=10}; bind any:all`,
		&[]struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`def any "foo" {x=10}; bind any`,
		&struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),
	rerror(`def any "foo" {x=10}; bind any:all`,
		&[]struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),

	// 30:
	rvalid(`def s "foo" {x=1}; bind s`, &S{}, &S{Name: "foo", X: 1}),
	rvalid(`def s "foo" {x=1}; bind s:all`, &[]S{}, &[]S{{Name: "foo", X: 1}}),
	rvalid(`def s {x=1}; def s{x=2}; bind s:all`, &[]S{}, &[]S{{X: 1}, {X: 2}}),
	rerror(`def y "foo" {x=1}; bind y`, &S{}, "mismatch: struct type S, block type y"),
	rerror(`def y "foo" {x=1}; bind y:all`, &[]S{}, "mismatch: struct type S, block type y"),

	rerror(`def s {y=1}; bind s`, &S{}, `field mapping for "y" not found in struct`),
	rerror(`def s {y=1}; bind s:all`, &[]S{}, `field mapping for "y" not found in struct`),
	rvalid(`def s2 "foo" {y=1}; bind s2`, &S2{}, &S2{Name: "foo", X: 1}),
	rvalid(`def s2 "foo" {y=1}; bind s2:all`, &[]S2{}, &[]S2{{Name: "foo", X: 1}}),
	rvalid(`def s2 "foo" {y=1}; bind s2:"foo",`, &[]S2{}, &[]S2{{Name: "foo", X: 1}}),

	// 40:
	rerror(`def s {x=""}; bind s`, &S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s {X=""}; bind s`, &S{},
		"type mismatch.+ struct.X has int, block.X has string",
	),
	rerror(`def s {x=""}; bind s:all`, &[]S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s "foo" {x=""}; bind s`, &S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s "foo" {x=""}; bind s:all`, &[]S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),

	// 45:
	rerror(`def s{}; def s{x=1}; bind s`, &S{}, "found 2 blocks of type s "),
	rerror(`def s{}; def s{x=1}; bind s:1`, &S{}, "found 2 blocks of type s "),
	rvalid(`def s{}; def s{x=1}; bind s:first`, &S{}, &S{}),
	rvalid(`def s{}; def s{x=1}; bind s:last `, &S{}, &S{X: 1}),
	rerror(`def s{}; def s{x=1}; bind s:all`, &S{}, "expected slice, have: struct"),

	// 50:
	rerror(`def s{}; def s{x=1}; bind s`, &S{}, "found 2 blocks of type s "),
	rerror(`def s{}; def s{x=1}; bind s:"foo"`, &S{}, `block s:"foo" not found`),
	rerror(`placeholder`, nil, "combined errors"),
	// ^^ rvalid(`def s{}; def s{x=1}; bind s:first,`, &[]S{}, &[]S{{}}),
	rerror(`placeholder`, nil, "combined errors"),
	// ^^ rvalid(`def s{}; def s{x=1}; bind s:last,  -> slice`, &[]S{}, &[]S{{X: 1}}),
	rvalid(`def s{}; def s{x=1}; bind s:all`, &[]S{}, &[]S{{}, {X: 1}}),

	// 55:
	rerror(`def foo{}; bind foo:2`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo:sth`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo:"q"`, &S{}, `block foo:"q" not found`),
	rerror(`def foo{}; bind "foo"`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo`, &S{}, "struct type S, block type foo"),

	// 60:
	rvalid(`def s3 "foo" {x=1}; bind s3 `, &S3{}, &S3{Name: "foo", X: 1}),
	rvalid(`def s3 "foo" {x=1}; bind s3`, &S3{}, &S3{X: 1, Name: "foo"}),
	rvalid(`def s4 {};          bind s4`, &S4{}, &S4{}),
	rvalid(`def s4 {x=1};       bind s4`, &S4{}, &S4{X: 1}),
	rvalid(`def s4 {x=1; y=2};  bind s4`, &S4{}, &S4{X: 1, Y: 2}),

	// 65:
	rerror(`def s {}; bind s:"foo"`, &S{}, `block s:"foo" not found`),
	rvalid(`def s "foo"{}; bind s:"foo"`, &S{}, &S{Name: "foo"}),
	rvalid(`def s "foo"{x=1}; bind s:"foo"`, &S{}, &S{Name: "foo", X: 1}),
	rvalid(`def s "foo"{x=1}; def s "bar"{x=2}; bind s:"foo"`, &S{},
		&S{Name: "foo", X: 1},
	),
	rvalid(`def s "foo"{x=1}; def s "bar"{x=2}; bind s:"bar"`, &S{},
		&S{Name: "bar", X: 2},
	),

	// 70:
	rvalid(`def s "foo"{}; bind s:"foo",`, &[]S{}, &[]S{{Name: "foo"}}),
	rerror(`def s "foo"{}; bind s:"foo2",`, &[]S{}, `block s:"foo2" not found`),
	rerror(`def s "foo"{}; bind s:"foo",`, &S{}, `expected slice, have: struct`),
	rvalid(`def s "foo"{}; def s "bar"{}; bind s:"foo","bar"`, &[]S{},
		&[]S{{Name: "foo"}, {Name: "bar"}},
	),
	rvalid(`def s "foo"{}; def s "bar"{}; bind s:"foo","bar",`, &[]S{},
		&[]S{{Name: "foo"}, {Name: "bar"}},
	),

	// 75:
	rvalid(`def s{x=5}; bind{s}`, &u{}, &u{S{X: 5}}),
	rvalid(`def s{x=5}; bind{s:1}`, &u{}, &u{S{X: 5}}),
	rvalid(`def s{x=5} def s{}; bind{s:first}`, &u{}, &u{S{X: 5}}),
	rvalid(`def s{x=5} def s{}; bind{s:last} `, &u{}, &u{S{X: 0}}),
	rvalid(`def s{x=5}; bind{s:all}`, &um{}, &um{[]S{{X: 5}}}),

	// 80:
	rerror(`def s{x=5}; bind{s}`, &[]int{}, `expected umbrella struct, have: slice`),
	rvalid(`def s "a"{}; bind{s:"a"}`, &u{}, &u{S{Name: "a"}}),
	rvalid(`def s "a"{}; bind{s:"a",}`, &um{}, &um{[]S{{Name: "a"}}}),
	rvalid(`def s "a"{}; def s "b"{}; bind{s:"a","b"}`, &um{},
		&um{[]S{{Name: "a"}, {Name: "b"}}},
	),
	rerror(`def y{x=5}; bind{y}`, &u{}, `mismatch: struct type S, block type y`),

	// 85:
	rerror(`def s{q=5}; bind{s}`, &u{}, `field mapping for "q" not found`),
	rerror(`def s{x=5}; bind{s}`, &struct{ x int }{}, `expected struct, have: int`),
	rvalid(`def s{x=5}; def s2{y=2}; bind{s; s2:all}`, &struct {
		S   S
		S2s []S2
	}{},
		&struct {
			S   S
			S2s []S2
		}{S{X: 5}, []S2{{X: 2}}},
	),
	rerror(`def s{x=5}; def s2{y=2}; bind{s; s2:all}`, &struct {
		S2s []S2
		S   S
	}{},
		`S2s: expected struct, have: slice`,
	),
	rerror(`def a{x=5}; bind{a}`, &struct{ A struct{ Bad int } }{},
		`A: field mapping for "x" not found`,
	),

	// 90
	rvalid(`def s5{}; bind s5`, &S5{}, &S5{}),
	rvalid(`def s5{x=2}; bind s5`, &S5{}, &S5{X: 2.0}),
	rvalid(`def s5{x=2.2}; bind s5`, &S5{}, &S5{X: 2.2}),
	rvalid(`def s5{i=3}; bind s5`, &S5{}, &S5{I: 3}),
	rerror(`def s5{i=3.2}; bind s5`, &S5{}, `type mismatch.+struct.I has int, block.i has float64`),
}

func TestReflect(t *testing.T) {
	for i, tc := range reflectTab {

		err := Unmarshal([]byte(tc.input), tc.dest, OptLogger(io.Discard))

		switch {
		case err != nil && tc.errp == nil:
			t.Errorf("tc#%d unexpected error: %v", i, err)

		case err != nil && tc.errp != nil:
			if !tc.errp.MatchString(err.Error()) {
				t.Errorf("tc#%d error mismatch\nhave: %v\nwant: %s",
					i, err, tc.errp,
				)
			}

		case err == nil && tc.errp != nil:
			t.Errorf("tc#%d have no error, want error pattern %q", i, tc.errp)

		default:
			have := tc.dest
			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf(
					"tc#%d mismatch:\nhave: %+v\nwant: %+v", i, have, tc.want,
				)
			}
		}
	}
}

func BenchmarkReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range reflectTab[:75] {
			_ = Unmarshal([]byte(tc.input), tc.dest, OptLogger(io.Discard))
		}
	}
}

var snakeTab = []struct{ input, want string }{
	{"", ""},
	{"x", "x"},
	{"X", "x"},
	{"foo", "foo"},
	{"Foo", "foo"},

	{"fooBar", "foo_bar"},
	{"FooBar", "foo_bar"},
	{"FooBAR", "foo_bar"},
	{"FOOBar", "foo_bar"},
	{"FOoBar", "f_oo_bar"},

	{"FoOBar", "fo_o_bar"},
	{"ab", "ab"},
	{"aB", "a_b"},
	{"Ab", "ab"},
	{"AB", "ab"},

	{"SNAKESEverywhere", "snakes_everywhere"},
	{"parseURLAndStuff", "parse_url_and_stuff"},
	{"ąćęŁŃÓŚŹŻ", "ąćę_łńóśźż"},
	{"ĄĆęŁŃÓŚŹŻ", "ą_ćę_łńóśźż"},
	{"ĄĆĘŁŃóśźż", "ąćęł_ńóśźż"},

	{"TCPDump", "tcp_dump"},
	{"SQLite", "sq_lite"}, // unfortunately
}

func TestSnakeUnsnake(t *testing.T) {

	for i, tc := range snakeTab {
		s := snake(tc.input)
		if s != tc.want {
			t.Errorf("tc#%d mismatch: have %q, want %q", i, s, tc.want)
		}
		if !unsnakeEq(tc.input, s) {
			t.Errorf("tc#%d unsnakeEq failed", i)
		}
	}
}

func BenchmarkSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range snakeTab[:20] {
			_ = snake(tc.input)
		}
	}
}
