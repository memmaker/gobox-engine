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

// TestOverviewColumns: the overview's level column comes after the name's
// full width; the columns keep their order, fit the page and take up the
// room a wide page has; the 40-column page keeps the old layout, no level.
func TestOverviewColumns(t *testing.T) {
	if c := overviewColumns(40); c != (overviewCols{hp: 19, ac: 27, status: 31}) {
		t.Errorf("40 columns: %+v, want the old page", c)
	}
	narrow, wide := overviewColumns(60), overviewColumns(120)
	for cols, c := range map[int]overviewCols{60: narrow, 120: wide} {
		if c.lv != 2+15+2 || !(c.lv < c.hp && c.hp < c.ac && c.ac < c.status && c.status < c.spells) {
			t.Errorf("%d columns: %+v, want the level right after the name, then HP, AC, Status, Spells", cols, c)
		}
		if last := max(c.spells+6, c.xp+10); last > cols-1 {
			t.Errorf("%d columns: %+v runs off the page", cols, c)
		}
	}
	if narrow.xp != 0 || wide.xp == 0 || wide.status-wide.ac <= narrow.status-narrow.ac {
		t.Errorf("60: %+v, 120: %+v: want XP only on the wide page, its columns further apart", narrow, wide)
	}
}
