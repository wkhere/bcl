package bcl

import "strings"

type lineCalc string

func (lc lineCalc) lineAt(pos int) int {
	return strings.Count(string(lc)[:pos], "\n") + 1
}
