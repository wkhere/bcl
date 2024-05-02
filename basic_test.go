package bcl_test

import (
	"testing"

	"github.com/wkhere/bcl"

	_ "embed"
)

//go:generate ./test.py generate

//go:embed testdata/basic_test.bcl
var basicInput []byte

func basicRun() ([]bcl.Block, error) {
	return bcl.Interpret(basicInput)
}

func TestBasic(t *testing.T) {
	_, err := basicRun()
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		basicRun()
	}
}
