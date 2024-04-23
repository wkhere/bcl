package bcl

import "testing"

func TestLineCalc(t *testing.T) {

	type lp struct{ l, p int }

	type p2lp = map[int]lp

	var tab = []struct {
		tid       string
		input     string
		positions p2lp
	}{
		{`empty`, "", p2lp{}},
		{`lf`, "\n", p2lp{0: lp{1, 1}}},
		{`1 char, no lf`, "a", p2lp{0: lp{1, 1}}},

		{`2 lines`, "foo\nbar\n", p2lp{
			0: lp{1, 1},
			1: lp{1, 2},
			2: lp{1, 3},
			3: lp{1, 4},
			4: lp{2, 1},
			6: lp{2, 3},
			7: lp{2, 4},
		}},
		{`2 lines, no ending lf`, "foo\nbar", p2lp{
			0: lp{1, 1},
			1: lp{1, 2},
			2: lp{1, 3},
			3: lp{1, 4},
			4: lp{2, 1},
			6: lp{2, 3},
			7: lp{2, 4},
		}},

		{`3 lines`, "1\n2\n3\n", p2lp{
			0: lp{1, 1}, 1: lp{1, 2},
			2: lp{2, 1}, 3: lp{2, 2},
			4: lp{3, 1}, 5: lp{3, 2},
		}},
	}

	for _, tc := range tab {
		lc := newLineCalc(tc.input)

		for wp, wlp := range tc.positions {
			l, p := lc.lineColAt(wp)
			if l != wlp.l || p != wlp.p {
				t.Errorf(
					"tc(%s) line[pos %d]: have %d:%d, want %d:%d",
					tc.tid, wp, l, p, wlp.l, wlp.p,
				)
			}
		}
	}
}
