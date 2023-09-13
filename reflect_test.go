package bcl

import (
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
