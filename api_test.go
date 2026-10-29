package bcl_test

//go:generate ./test.py generate testapi_test.go

import (
	"errors"
	"io"
	"testing"
	"time"

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

func TestParseFileEmptyChunk(t *testing.T) {
	// empty chunk at a token boundary used to be taken as eof,
	// with the reader left blocked, so ParseFile hung
	r := &chunkfile{chunks: []string{"var a = 1;", "", "def x { y = a }"}}
	type res struct {
		blocks []bcl.Block
		err    error
	}
	resc := make(chan res, 1)
	go func() {
		blocks, _, err := bcl.InterpretFile(r)
		resc <- res{blocks, err}
	}()
	select {
	case res := <-resc:
		if res.err != nil {
			t.Fatal(res.err)
		}
		if len(res.blocks) != 1 {
			t.Errorf("input truncated: got %d blocks, want 1", len(res.blocks))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ParseFile hangs")
	}
}

func TestParseFileSplitRune(t *testing.T) {
	// a rune split across chunks used to be decoded as two RuneErrors;
	// here a non-breaking space, which the lexer must see whole
	src := "var a = 1 eval a"
	i := len("var a = 1") + 1

	for _, chunks := range [][]string{
		{src[:i], src[i:]},
		{src[:i], "", src[i : i+1], src[i+1:]}, // empty chunk inside a rune
		bytewise(src),
	} {
		_, err := bcl.ParseFile(&chunkfile{chunks: chunks})
		if err != nil {
			t.Errorf("%q: %v", chunks, err)
		}
	}
}

func bytewise(s string) []string {
	cc := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		cc[i] = s[i : i+1]
	}
	return cc
}

type chunkfile struct{ chunks []string }

func (c *chunkfile) Read(p []byte) (int, error) {
	if len(c.chunks) == 0 {
		return 0, io.EOF
	}
	n := copy(p, c.chunks[0])
	c.chunks = c.chunks[1:]
	return n, nil
}

func (*chunkfile) Close() error { return nil }
func (*chunkfile) Name() string { return "<chunks>" }
