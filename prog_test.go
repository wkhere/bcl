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

func TestBasicProgDumpLoad(t *testing.T) {
	testDumpLoad(basicInput, t)
}
