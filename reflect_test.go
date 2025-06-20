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

	bind struct1 -> struct
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

var reflectTab = []reflecttc{

	rerror(``, nil, "no binding"),
	rerror(``, 1, "no binding"),
	rerror(``, struct{}{}, "no binding"),
	rerror(``, &struct{}{}, "no binding"),
	rerror(``, &[]struct{}{}, "no binding"),

	rerror(`bind any -> struct`, nil, "no blocks of type any"),
	rerror(`bind any -> struct`, 1, "no blocks of type any"),
	rerror(`bind any -> struct`, struct{}{}, "no blocks of type any"),
	rerror(`bind any -> struct`, &struct{}{}, "no blocks of type any"),
	rerror(`bind any -> struct`, &[]struct{}{}, "no blocks of type any"),

	rerror(`def any{}; bind any -> struct`, nil, "expected pointer"),
	rerror(`def any{}; bind any -> struct`, 1, "expected pointer, have: int"),
	rerror(`def any{}; bind any -> struct`, struct{}{}, "expected pointer, have: struct"),
	rvalid(`def any{}; bind any -> struct`, &struct{}{}, &struct{}{}),
	rerror(`def any{}; bind any -> struct`, &[]struct{}{},
		"pointer deref: expected struct, have: slice",
	),

	rerror(`def any{}; bind any -> slice`, nil, "expected pointer"),
	rerror(`def any{}; bind any -> slice`, struct{}{}, "expected pointer, have: struct"),
	rerror(`def any{}; bind any -> slice`, &struct{}{},
		"pointer deref: expected slice, have: struct",
	),
	rvalid(`def any{}; bind any -> slice`, &[]struct{}{}, &[]struct{}{{}}), // empty struct inside
	// the next one is nasty: would panic if there was no checking for a slice element:
	rerror(`def int{}; bind int -> slice`, &[]int{},
		"slice element deref: expected struct, have: int",
	),

	rvalid(`def any{x=1}; bind any -> struct`, &struct{ X int }{}, &struct{ X int }{X: 1}),
	rvalid(`def any{x=1}; bind any:all -> slice`, &[]struct{ X int }{}, &[]struct{ X int }{{X: 1}}),
	rvalid(`def any "foo" {}; bind any -> struct`,
		&struct{ Name string }{},
		&struct{ Name string }{Name: "foo"},
	),
	rvalid(`def any "foo" {}; bind any:all -> slice`,
		&[]struct{ Name string }{},
		&[]struct{ Name string }{{Name: "foo"}},
	),
	rerror(`def any "foo" {}; bind any -> struct`,
		&struct{}{},
		`field mapping for "Name" not found in struct`,
	),

	rvalid(`def any "foo" {s="quux"}; bind any -> struct`,
		&struct{ Name, S string }{},
		&struct{ Name, S string }{Name: "foo", S: "quux"},
	),
	rerror(`def any "foo" {x=10}; bind any -> struct`,
		&struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`def any "foo" {x=10}; bind any:all -> slice`,
		&[]struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`def any "foo" {x=10}; bind any -> struct`,
		&struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),
	rerror(`def any "foo" {x=10}; bind any -> slice`,
		&[]struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),

	rvalid(`def s "foo" {x=1}; bind s -> struct`, &S{}, &S{Name: "foo", X: 1}),
	rvalid(`def s "foo" {x=1}; bind s -> slice`, &[]S{}, &[]S{{Name: "foo", X: 1}}),
	rvalid(`def s {x=1}; def s{x=2}; bind s:all -> slice`, &[]S{}, &[]S{{X: 1}, {X: 2}}),
	rerror(`def y "foo" {x=1}; bind y -> struct`, &S{}, "mismatch: struct type S, block type y"),
	rerror(`def y "foo" {x=1}; bind y -> slice`, &[]S{}, "mismatch: struct type S, block type y"),

	rerror(`def s {y=1}; bind s -> struct`, &S{}, `field mapping for "y" not found in struct`),
	rerror(`def s {y=1}; bind s -> slice`, &[]S{}, `field mapping for "y" not found in struct`),
	rvalid(`def s2 "foo" {y=1}; bind s2 -> struct`, &S2{}, &S2{Name: "foo", X: 1}),
	rvalid(`def s2 "foo" {y=1}; bind s2 -> slice`, &[]S2{}, &[]S2{{Name: "foo", X: 1}}),
	rvalid(`def s2 "foo" {y=1}; bind s2:all -> slice`, &[]S2{}, &[]S2{{Name: "foo", X: 1}}),

	rerror(`def s {x=""}; bind s -> struct`, &S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s {X=""}; bind s -> struct`, &S{},
		"type mismatch.+ struct.X has int, block.X has string",
	),
	rerror(`def s {x=""}; bind s -> slice`, &[]S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s "foo" {x=""}; bind s -> struct`, &S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
	rerror(`def s "foo" {x=""}; bind s -> slice`, &[]S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),

	rerror(`def s{}; def s{x=1}; bind s -> struct`, &S{}, "found 2 blocks of type s "),
	rerror(`def s{}; def s{x=1}; bind s:1 -> struct`, &S{}, "found 2 blocks of type s "),
	rvalid(`def s{}; def s{x=1}; bind s:first -> struct`, &S{}, &S{}),
	rvalid(`def s{}; def s{x=1}; bind s:last  -> struct`, &S{}, &S{X: 1}),
	rerror(`def s{}; def s{x=1}; bind s:all   -> struct`, &S{}, "combined errors from parse"),

	rerror(`def s{}; def s{x=1}; bind s -> slice`, &[]S{}, "found 2 blocks of type s "),
	rerror(`def s{}; def s{x=1}; bind s -> slice`, &[]S{}, "found 2 blocks of type s "),
	rvalid(`def s{}; def s{x=1}; bind s:first -> slice`, &[]S{}, &[]S{{}}),
	rvalid(`def s{}; def s{x=1}; bind s:last  -> slice`, &[]S{}, &[]S{{X: 1}}),
	rvalid(`def s{}; def s{x=1}; bind s:all   -> slice`, &[]S{}, &[]S{{}, {X: 1}}),

	rerror(`def foo{}; bind foo:2   -> struct`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo:sth -> struct`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo:"q" -> struct`, &S{}, `block foo:"q" not found`),
	rerror(`def foo{}; bind "foo"   -> struct`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo     -> oopsie`, &S{}, "combined errors from parse"),

	rvalid(`def s3 "foo" {x=1}; bind s3 -> struct`, &S3{}, &S3{Name: "foo", X: 1}),
	rvalid(`def s3 "foo" {x=1}; bind s3 -> struct`, &S3{}, &S3{X: 1, Name: "foo"}),
	rvalid(`def s4 {};         bind s4 -> struct`, &S4{}, &S4{}),
	rvalid(`def s4 {x=1};      bind s4 -> struct`, &S4{}, &S4{X: 1}),
	rvalid(`def s4 {x=1; y=2}; bind s4 -> struct`, &S4{}, &S4{X: 1, Y: 2}),

	rerror(`def s {}; bind s:"foo" -> struct`, &S{}, `block s:"foo" not found`),
	rvalid(`def s "foo"{}; bind s:"foo" -> struct`, &S{}, &S{Name: "foo"}),
	rvalid(`def s "foo"{x=1}; bind s:"foo" -> struct`, &S{}, &S{Name: "foo", X: 1}),
	rvalid(`def s "foo"{x=1}; def s "bar"{x=2}; bind s:"foo" -> struct`, &S{},
		&S{Name: "foo", X: 1},
	),
	rvalid(`def s "foo"{x=1}; def s "bar"{x=2}; bind s:"bar" -> struct`, &S{},
		&S{Name: "bar", X: 2},
	),

	rvalid(`def s "foo"{}; bind s:"foo"  -> slice`, &[]S{}, &[]S{{Name: "foo"}}),
	rerror(`def s "foo"{}; bind s:"foo2" -> slice`, &[]S{}, `block s:"foo2" not found`),
	rerror(`def s "foo"{}; bind s:"foo", -> struct`, &S{}, `invalid bind target and selector`),
	rvalid(`def s "foo"{}; bind s:"foo", -> slice`, &[]S{}, &[]S{{Name: "foo"}}),
	rvalid(`def s "foo"{}; def s "bar"{}; bind s:"foo","bar" -> slice`, &[]S{},
		&[]S{{Name: "foo"}, {Name: "bar"}},
	),
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
