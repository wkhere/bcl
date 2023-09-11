package bcl

import (
	"testing"

	_ "embed"
)

//go:embed test.bcl
var basicInput []byte

func basicRun() ([]Block, error) {
	return Interp(basicInput)
}

func TestBasic(t *testing.T) {
	_, err := basicRun()
	if err != nil {
		t.Error("unexpected error:", err)
	}
}

func BenchmarkBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		basicRun()
	}
}
