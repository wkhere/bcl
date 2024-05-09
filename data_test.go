package bcl_test

import (
	"bytes"
	"testing"

	"github.com/wkhere/bcl"

	_ "embed"
)

//go:embed testdata/basic_test.bcl
var basicInput []byte

//go:embed testdata/big1.bcl
var big1 []byte

func TestBasicBytes(t *testing.T) {
	_, err := bcl.Interpret(basicInput)
	if err != nil {
		t.Error(err)
	}
}

func TestBasicFile(t *testing.T) {
	r := reader{bytes.NewReader(basicInput)}
	_, err := bcl.InterpretFile(r)
	if err != nil {
		t.Error(err)
	}
}

func TestTokenLeftover(t *testing.T) {
	r := reader{bytes.NewReader(big1)}
	_, err := bcl.ParseFile(r)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBasicBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bcl.Interpret(basicInput)
	}
}

func BenchmarkBasicFile(b *testing.B) {
	r := reader{bytes.NewReader(basicInput)}
	for i := 0; i < b.N; i++ {
		r.Seek(0, 0)
		bcl.InterpretFile(r)
	}
}

type reader struct{ *bytes.Reader }

func (reader) Close() error { return nil }
func (reader) Name() string { return "*test-reader*" }
