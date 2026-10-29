package bcl_test

import (
	"bytes"
	"fmt"
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

func testLoadTruncated(input []byte, t *testing.T) {
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
	full := b.Bytes()

	load := func(p []byte) (err error) {
		defer func() {
			if x := recover(); x != nil {
				err = fmt.Errorf("panic: %v", x)
			}
		}()
		_, err = bcl.LoadProg(bytes.NewReader(p), "input", bcl.OptOutput(io.Discard))
		return err
	}

	for i := range full {
		err := load(full[:i])
		switch {
		case err == nil:
			t.Errorf("truncated at %d/%d: no error", i, len(full))
		case strings.HasPrefix(err.Error(), "panic:"):
			t.Errorf("truncated at %d/%d: %v", i, len(full), err)
		}
	}
}

// varintsInput has multi-byte uvarints in various places:
// int constant, string length, and enough code for large positions.
var varintsInput = []byte(`
var i = 123456789
var f = 3.14
var s = "` + strings.Repeat("x", 300) + `"
def blk "name" { x = i; y = f; z = s; w = true; n = nil }
`)

func TestBasicDumpLoad(t *testing.T) {
	testDumpLoad(basicInput, t)
}

func TestDumpWithTrailingData(t *testing.T) {
	testDumpWithTrailingData(basicInput, t)
}

func TestLoadTruncated(t *testing.T) {
	testLoadTruncated(basicInput, t)
	testLoadTruncated(varintsInput, t)
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
