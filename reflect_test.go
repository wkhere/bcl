package bcl

import (
	"reflect"
	"strings"
	"testing"
)

type Struct1 struct {
	Name         string
	Field1       int
	Field2       string
	Status       bool `bcl:"field3"`
	AnotherField int
	Other        int `foo:"bar" bcl:"other_field"`
}

func simpleUnmarshal(dest any) error {
	const bcl = `
	struct1 "foo" {
		field1 = 10
		field2 = "abc"
		field3 = true
		another_field = 10
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
	errs  string
}

type rr = []Struct1

func rvalid(in string, d, w any) reflecttc {
	return reflecttc{in, d, w, ""}
}
func rerror(in string, d any, e string) reflecttc {
	return reflecttc{in, d, nil, e}
}

type S struct {
	Name string
	X    int
}

var reflectTab = []reflecttc{
	rvalid(``, &rr{}, &rr{}),

	rerror(``, nil, "expected pointer to a slice of structs"),
	rerror(``, 1, "expected pointer to a slice of structs"),
	rerror(``, "foo", "expected pointer to a slice of structs"),
	rerror(``, struct{}{}, "expected pointer to a slice of structs"),
	rerror(``, &struct{}{}, "expected pointer to a slice of structs"),
	rerror(``, &struct{}{}, "expected pointer to a slice of structs"),

	rvalid(`any "foo"{}`,
		&[]struct{ Name string }{},
		&[]struct{ Name string }{{Name: "foo"}},
	),
	rerror(`any "foo"{x=10}`,
		&[]struct{ Name string }{},
		`field mapping for "x" not found in struct`,
	),
	rerror(`any "foo"{x=10}`,
		&[]struct {
			Name string
			x    int
		}{},
		`found field "x" but is unexported`,
	),

	rerror(``, []S{}, "expected pointer to a slice of structs"),

	rvalid(`s "foo"{x=10}`, &[]S{}, &[]S{{Name: "foo", X: 10}}),
	rerror(`y "foo"{x=10}`, &[]S{}, "mismatch: struct type S, block type y"),
}

func TestReflect(t *testing.T) {
	for i, tc := range reflectTab {

		err := Unmarshal([]byte(tc.input), tc.dest)

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
			have := tc.dest
			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf(
					"tc#%d mismatch:\nhave %+v\nwant %+v", i, have, tc.want,
				)
			}
		}
	}
}
