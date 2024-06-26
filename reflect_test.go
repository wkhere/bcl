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
	}`
	return Unmarshal([]byte(bcl), dest)
}

func TestSimpleUnmarshal(t *testing.T) {
	var a []Struct1

	err := simpleUnmarshal(&a)
	if err != nil {
		t.Error(err)
	}
	x := a[0]
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
	t.Logf("%+v", a)
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

type rr = []Struct1

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
	rvalid(``, &rr{}, &rr{}),

	rerror(``, nil, "expected pointer to a slice of structs"),
	rerror(``, 1, "expected pointer to a slice of structs"),
	rerror(``, "foo", "expected pointer to a slice of structs"),
	rerror(``, struct{}{}, "expected pointer to a slice of structs"),
	rerror(``, &struct{}{}, "expected pointer to a slice of structs"),
	rerror(``, &struct{}{}, "expected pointer to a slice of structs"),

	rvalid(`def any "foo"{}`,
		&[]struct{ Name string }{},
		&[]struct{ Name string }{{Name: "foo"}},
	),
	rerror(`def any "foo"{}`,
		&[]struct{}{},
		`field mapping for "Name" not found in struct`,
	),
	rvalid(`def any {}`,
		&[]struct{}{},
		&[]struct{}{{}}, // note there is struct inside, just empty
	),
	rvalid(`def any {x=1}`,
		&[]struct{ X int }{},
		&[]struct{ X int }{{X: 1}},
	),
	rerror(`def any "foo"{x=10}`,
		&[]struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`def any "foo"{x=10}`,
		&[]struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),

	rerror(``, []S{}, "expected pointer to a slice of structs"),

	rvalid(`def s "foo"{x=1}`, &[]S{}, &[]S{{Name: "foo", X: 1}}),
	rerror(`def y "foo"{x=1}`, &[]S{}, "mismatch: struct type S, block type y"),
	rerror(`def s "foo"{y=1}`, &[]S{}, `field mapping for "y" not found in struct`),
	rvalid(`def s2 "foo"{y=1}`, &[]S2{}, &[]S2{{Name: "foo", X: 1}}),
	rerror(`def s "foo"{x=""}`, &[]S{},
		"type mismatch.+ struct.X has int, block.x has string",
	),
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
