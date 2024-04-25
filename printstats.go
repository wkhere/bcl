package bcl

import "fmt"

func printPStats(pstats parseStats) {
	fmt.Printf("pstats.tokens:     %5d\n", pstats.tokens)
	fmt.Printf("pstats.localMax:   %5d\n", pstats.localMax)
	fmt.Printf("pstats.depthMax:   %5d\n", pstats.depthMax)
	fmt.Printf("pstats.constants:  %5d\n", pstats.constants)
	fmt.Printf("pstats.opsCreated: %5d\n", pstats.opsCreated)
	fmt.Printf("pstats.codeBytes:  %5d\n", pstats.codeBytes)
}

func printXStats(xstats execStats) {
	fmt.Printf("xstats.tosMax:     %5d\n", xstats.tosMax)
	fmt.Printf("xstats.blockTosMax:%5d\n", xstats.blockTosMax)
	fmt.Printf("xstats.opsRead:    %5d\n", xstats.opsRead)
	fmt.Printf("xstats.pcFinal:    %5d\n", xstats.pcFinal)
}
