package ui

import "testing"

// TestNoKeyRows: a NoKey row gets no digit, so a list of them draws every
// row, greyed or not, at the same column.
func TestNoKeyRows(t *testing.T) {
	do := func() {}
	rows := []MenuRow{{Text: "Cure", NoKey: true, Do: do}, {Text: "Heal", NoKey: true}, {Text: "Exit", Key: 'E', Do: do}}
	keys := RowKeys(rows)
	if keys[0] != 0 || keys[1] != 0 || keys[2] != 'E' {
		t.Fatalf("keys %q", keys)
	}
	for i, r := range rows[:2] {
		if s, _ := HotText(r.Text, keys[i]); s != r.Text {
			t.Errorf("row %d drawn %q", i, s)
		}
	}
}
