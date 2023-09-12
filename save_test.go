package bcl

import (
	"testing"
)

func TestSimpleSave(t *testing.T) {
	const bcl = `
	struct1 "foo" {
		field1 = 10
		field2 = "abc"
		field3 = true
	}`

	type Struct1 struct {
		Name   string
		Field1 int
		Field2 string
		Field3 bool
	}
	var a []Struct1

	err := InterpAndSave(&a, []byte(bcl))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", a)
}
