package ui

import "testing"

// TestThousands: the purse's numbers read with a comma between thousands.
func TestThousands(t *testing.T) {
	for n, want := range map[int]string{0: "0", 37: "37", 999: "999", 1000: "1,000", 51211: "51,211", 1234567: "1,234,567", -4500: "-4,500"} {
		if got := thousands(n); got != want {
			t.Errorf("thousands(%d) = %q, want %q", n, got, want)
		}
	}
}
