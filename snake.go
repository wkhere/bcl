package bcl

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func unsnakeMatcher(snake string) func(string) bool {
	u := strings.ReplaceAll(snake, "_", "")
	return func(s string) bool {
		return strings.EqualFold(s, u)
	}
}

func unsnakeEq(orig, snake string) bool {
	return unsnakeMatcher(snake)(orig)
}

func snake(input string) string {
	nextRune := func(idx int) rune { r, _ := utf8.DecodeRuneInString(input[idx:]); return r }

	var b strings.Builder
	var prev rune

	for i, v := range input {
		if unicode.IsUpper(v) {
			if i > 0 && (unicode.IsLower(prev) ||
				unicode.IsLower(nextRune(i+utf8.RuneLen(v)))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(v))
		} else {
			b.WriteRune(v)
		}
		prev = v
	}
	return b.String()
}
