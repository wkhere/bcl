package bcl

import (
	"fmt"
	"os"
)

type logger struct {
	w *os.File
}

var log = logger{os.Stderr}

func (l logger) Print(a ...any)                 { fmt.Fprint(l.w, a...) }
func (l logger) Println(a ...any)               { fmt.Fprintln(l.w, a...) }
func (l logger) Printf(format string, a ...any) { fmt.Fprintf(l.w, format, a...) }
