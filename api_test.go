package bcl_test

//go:generate ./test.py generate testapi_test.go

import (
	"errors"
	"testing"

	"github.com/wkhere/bcl"
)

func TestParseFileErr(t *testing.T) {
	err1 := errors.New("test err")
	r := &errfile{err: err1}
	_, err := bcl.ParseFile(r)
	if err != err1 {
		t.Errorf("errors mismatch")
	}
}

type errfile struct {
	first bool
	err   error
}

func (e *errfile) Read(p []byte) (int, error) {
	if e.first {
		e.first = false
		return copy(p, "foo\n"), e.err
	}
	return 0, e.err
}

func (*errfile) Close() error { return nil }
func (*errfile) Name() string { return "<err>" }
