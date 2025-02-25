package bcl

import (
	"testing"
)

var snakeTab = []struct{ input, want string }{
	{"", ""},
	{"x", "x"},
	{"X", "x"},
	{"foo", "foo"},
	{"Foo", "foo"},

	{"fooBar", "foo_bar"},
	{"FooBar", "foo_bar"},
	{"FooBAR", "foo_bar"},
	{"FOOBar", "foo_bar"},
	{"FOoBar", "f_oo_bar"},

	{"FoOBar", "fo_o_bar"},
	{"ab", "ab"},
	{"aB", "a_b"},
	{"Ab", "ab"},
	{"AB", "ab"},

	{"SNAKESEverywhere", "snakes_everywhere"},
	{"parseURLAndStuff", "parse_url_and_stuff"},
	{"ąćęŁŃÓŚŹŻ", "ąćę_łńóśźż"},
	{"ĄĆęŁŃÓŚŹŻ", "ą_ćę_łńóśźż"},
	{"ĄĆĘŁŃóśźż", "ąćęł_ńóśźż"},
}

func TestSnakeUnsnake(t *testing.T) {

	for i, tc := range snakeTab {
		s := snake(tc.input)
		if s != tc.want {
			t.Errorf("tc#%d mismatch: have %q, want %q", i, s, tc.want)
		}
		if !unsnakeEq(tc.input, s) {
			t.Errorf("tc#%d unsnakeEq failed", i)
		}
	}
}

func BenchmarkSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range snakeTab[:20] {
			_ = snake(tc.input)
		}
	}
}
