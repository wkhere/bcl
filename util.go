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

type writers struct {
	outw, logw io.Writer
}

func (pstats *parseStats) print(w io.Writer) {
	fmt.Fprintf(w, "pstats.tokens:     %5d\n", pstats.tokens)
	fmt.Fprintf(w, "pstats.localMax:   %5d\n", pstats.localMax)
	fmt.Fprintf(w, "pstats.depthMax:   %5d\n", pstats.depthMax)
	fmt.Fprintf(w, "pstats.constants:  %5d\n", pstats.constants)
	fmt.Fprintf(w, "pstats.opsCreated: %5d\n", pstats.opsCreated)
	fmt.Fprintf(w, "pstats.codeBytes:  %5d\n", pstats.codeBytes)
}

func (xstats *execStats) print(w io.Writer) {
	fmt.Fprintf(w, "xstats.tosMax:     %5d\n", xstats.tosMax)
	fmt.Fprintf(w, "xstats.blockTosMax:%5d\n", xstats.blockTosMax)
	fmt.Fprintf(w, "xstats.opsRead:    %5d\n", xstats.opsRead)
	fmt.Fprintf(w, "xstats.pcFinal:    %5d\n", xstats.pcFinal)
}
