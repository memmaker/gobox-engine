package rules

import (
	"strings"

	"gobox/core/goldbox"
)

// MakeFileName is the DOS file name of a character name: the separators are
// removed, at most 8 characters are kept, upper case.
func MakeFileName(s string) string {
	for _, c := range " .*,?/\\:;|" {
		s = goldbox.RemoveSubstr(s, string(c))
	}
	if len(s) > 8 {
		s = s[:8]
	}
	return strings.ToUpper(s)
}
