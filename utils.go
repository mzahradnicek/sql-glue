package sqlg

import "strings"

func ToSnake(camel string) string {
	var b strings.Builder
	l := len(camel) - 1

	for i, v := range camel {
		// A is 65, a is 97
		if v >= 'a' || v < 'A' {
			b.WriteRune(v)
			continue
		}
		// v is capital letter here
		// irregard first letter
		// add underscore if last letter is capital letter
		// add underscore when previous letter is lowercase
		// add underscore when next letter is lowercase
		if i != 0 && ((rune(camel[i-1]) >= 'a' || rune(camel[i-1]) < 'A') || // pre
			(i < l && rune(camel[i+1]) >= 'a')) { //next
			b.WriteRune('_')
		}
		b.WriteRune(v + 32) // 'a'-'A' = 32
	}

	return b.String()
}
