package bcl

import (
	"fmt"
	"sort"
)

type lineCalc struct {
	lfs []int
}

func newLineCalc(s string) *lineCalc {
	lfs := make([]int, 0, len(s)/32+4)
	for i, c := range s {
		if c == '\n' {
			lfs = append(lfs, i)
		}
	}
	return &lineCalc{lfs: lfs}
}

// lineColAt gives (line, column) pair for a given position.
// Note that pos starts at 0, while line and column start at 1.
func (lc *lineCalc) lineColAt(pos int) (int, int) {

	j := sort.SearchInts(lc.lfs, pos)

	if j == len(lc.lfs) {
		if j == 0 {
			return 1, pos + 1
		}
		return j + 1, pos - lc.lfs[j-1]
	}

	prevpos := -1
	if j > 0 {
		prevpos = lc.lfs[j-1]
	}
	return j + 1, pos - prevpos
}

func (lc *lineCalc) lineAt(pos int) int {
	line, _ := lc.lineColAt(pos)
	return line
}

func (lc *lineCalc) format(pos int) string {
	l, p := lc.lineColAt(pos)
	return fmt.Sprintf("%d:%d", l, p)
}
