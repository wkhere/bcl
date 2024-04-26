package bcl

import (
	"fmt"
	"io"
)

func printPStats(w io.Writer, pstats parseStats) {
	fmt.Fprintf(w, "pstats.tokens:     %5d\n", pstats.tokens)
	fmt.Fprintf(w, "pstats.localMax:   %5d\n", pstats.localMax)
	fmt.Fprintf(w, "pstats.depthMax:   %5d\n", pstats.depthMax)
	fmt.Fprintf(w, "pstats.constants:  %5d\n", pstats.constants)
	fmt.Fprintf(w, "pstats.opsCreated: %5d\n", pstats.opsCreated)
	fmt.Fprintf(w, "pstats.codeBytes:  %5d\n", pstats.codeBytes)
}

func printXStats(w io.Writer, xstats execStats) {
	fmt.Fprintf(w, "xstats.tosMax:     %5d\n", xstats.tosMax)
	fmt.Fprintf(w, "xstats.blockTosMax:%5d\n", xstats.blockTosMax)
	fmt.Fprintf(w, "xstats.opsRead:    %5d\n", xstats.opsRead)
	fmt.Fprintf(w, "xstats.pcFinal:    %5d\n", xstats.pcFinal)
}
