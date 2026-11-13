package rules

import "testing"

// Pinned on the games' copies: the scan does not recheck after a deletion,
// so "aaaa" loses every second "a" only.
func TestUtilPinned(t *testing.T) {
	if got := RemoveSubstr("aaaa", "a"); got != "aa" {
		t.Errorf("RemoveSubstr = %q, want aa", got)
	}
	if Abs(-3) != 3 || Sign(-9) != -1 || Sign(0) != 0 || Cut("abcd", 2) != "ab" || Upper('q') != 'Q' {
		t.Error("helper wrong")
	}
}
