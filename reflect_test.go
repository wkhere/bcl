package bcl

import (
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
	var a []Struct1

	for i := 0; i < b.N; i++ {
		simpleUnmarshal(&a)
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
	rerror(`def any{}; bind any -> struct`, 1, "expected pointer"),
	rerror(`def any{}; bind any -> struct`, struct{}{}, "expected pointer"),
	rvalid(`def any{}; bind any -> struct`, &struct{}{}, &struct{}{}),
	rerror(`def any{}; bind any -> struct`, &[]struct{}{}, "expected pointer to a struct"),

	rerror(`def any{}; bind any -> slice`, nil, "expected pointer"),
	rerror(`def any{}; bind any -> slice`, 1, "expected pointer"),
	rerror(`def any{}; bind any -> slice`, struct{}{}, "expected pointer"),
	rerror(`def any{}; bind any -> slice`, &struct{}{}, "expected pointer to a slice of structs"),
	rvalid(`def any{}; bind any -> slice`, &[]struct{}{}, &[]struct{}{{}}), // empty struct inside

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
	rerror(`def foo{}; bind foo:"q" -> struct`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind "foo"   -> struct`, &S{}, "combined errors from parse"),
	rerror(`def foo{}; bind foo     -> oopsie`, &S{}, "combined errors from parse"),
}

func TestReflect(t *testing.T) {
	for i, tc := range reflectTab {

		err := Unmarshal([]byte(tc.input), tc.dest)

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
