package bcl_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/wkhere/bcl"
)

func testDumpLoad(input []byte, t *testing.T) {
	t.Helper()

	prog, err := bcl.Parse(input, "input", bcl.OptOutput(io.Discard))
	if err != nil {
		t.Errorf("parse error: %v", err)
		return
	}

	b := new(bytes.Buffer)
	err = prog.Dump(b)
	if err != nil {
		t.Errorf("dump error: %v", err)
	}

	prog2, err := bcl.LoadProg(b, "input", bcl.OptOutput(io.Discard))
	if err != nil {
		t.Errorf("load error: %v", err)
	}

	if !reflect.DeepEqual(prog2, prog) {
		t.Errorf("progs not equal")
	}
}

func TestBasicDumpLoad(t *testing.T) {
	testDumpLoad(basicInput, t)
}

func benchDumpLoad(input []byte, b *testing.B) {
	prog, _ := bcl.Parse(input, "input", bcl.OptOutput(io.Discard))

	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		prog.Dump(buf)
		bcl.LoadProg(buf, "input", bcl.OptOutput(io.Discard))
	}
}

func BenchmarkBasicDumpLoad(b *testing.B) {
	benchDumpLoad(basicInput, b)
}
