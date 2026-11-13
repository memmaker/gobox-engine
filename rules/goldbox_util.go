package rules

// Abs is |n|.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Sign is -1, 0 or 1.
func Sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// Cut limits s to n bytes.
func Cut(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// Upper is the original's UpCase.
func Upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

// RemoveSubstr removes every copy of sub from s.
// The scan runs over the original length and does not recheck the position
// after a deletion, so adjacent copies leave every second one (original quirk).
func RemoveSubstr(s, sub string) string {
	if sub == "" {
		return s
	}
	for i, n := 1, len(s)-len(sub)+1; i <= n; i++ {
		if i-1+len(sub) <= len(s) && s[i-1:i-1+len(sub)] == sub {
			s = s[:i-1] + s[i-1+len(sub):]
		}
	}
	return s
}
