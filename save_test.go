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
		another_field = 10
		other_field = 42
	}`

	type Struct1 struct {
		Name         string
		Field1       int
		Field2       string
		Status       bool `bcl:"field3"`
		AnotherField int
		Other        int `foo:"bar" bcl:"other_field"`
	}
	var a []Struct1

	err := Unmarshal([]byte(bcl), &a)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", a)
}
