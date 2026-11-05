package ui

import "testing"

func TestDigitPad(t *testing.T) {
	rows := make([]MenuRow, 12)
	for i := range rows {
		rows[i] = MenuRow{Text: "Cure", Do: func() {}}
	}
	rows[11].Key = 'E'
	keys := RowKeys(rows)
	items := make([]string, len(rows))
	for i := range items {
		items[i], _ = HotText("Cure", keys[i])
	}
	digitPad(items, keys, nil)
	for i := range 11 {
		if len(items[i]) != len(items[0]) {
			t.Errorf("row %d %q not under %q", i, items[i], items[0])
		}
	}
}
