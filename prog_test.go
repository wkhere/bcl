package bcl_test

import (
	"bytes"
	"io"
	"reflect"
	"strings"
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
		return
	}

	prog2, err := bcl.LoadProg(b, "input", bcl.OptOutput(io.Discard))
	if err != nil {
		t.Errorf("load error: %v", err)
		return
	}

	if !reflect.DeepEqual(prog2, prog) {
		t.Errorf("progs not equal")
	}
}

func testDumpWithTrailingData(input []byte, t *testing.T) {
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
		return
	}
	b.WriteByte(0)

	_, err = bcl.LoadProg(b, "input", bcl.OptOutput(io.Discard))
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Errorf("error should match `trailing data`, have: %v", err)
	}
}

func TestBasicDumpLoad(t *testing.T) {
	testDumpLoad(basicInput, t)
}

func TestDumpWithTrailingData(t *testing.T) {
	testDumpWithTrailingData(basicInput, t)
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
