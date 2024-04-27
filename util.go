package bcl

import (
	"fmt"
	"io"
)

// logger is a good enough solution for parser error logs.
// Stdlib log.Logger features, like support for concurrent writes, is not needed
// here, as the parser is sequential.
type logger struct {
	w io.Writer
}

func (l logger) Print(a ...any)                 { fmt.Fprint(l.w, a...) }
func (l logger) Println(a ...any)               { fmt.Fprintln(l.w, a...) }
func (l logger) Printf(format string, a ...any) { fmt.Fprintf(l.w, format, a...) }
