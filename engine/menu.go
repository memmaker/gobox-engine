package engine

import "strings"

// Menu words (ovr26's menu bars): a hotkey is [0-9A-Z]; the questions
// they make are eclask.go's.

func isHotkey(c byte) bool { return c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' } // set [0-9A-Z]

// optionText (extra): how a word reads as an option. A hotkey inside a
// word reads lower case and the word's first letter upper case, so "sWeep"
// reads "Sweep", "aiM" "Aim".
func optionText(w string) string {
	b := []byte(w)
	lower := func(c byte) bool { return c >= 'a' && c <= 'z' }
	hasLower := strings.IndexFunc(w, func(r rune) bool { return r >= 'a' && r <= 'z' }) >= 0
	for i, c := range b {
		switch {
		case i == 0 && lower(c):
			b[i] = c - 32
		case i > 0 && hasLower && c >= 'A' && c <= 'Z':
			b[i] = c + 32
		}
	}
	return string(b)
}
