package combat

import "testing"

func TestShownItemsLabels(t *testing.T) {
	b := &Board{Turn: &Turn{Can: []Kind{ActStep, ActView, ActAim, ActTurn, ActQuick, ActEnd, Kind(999)},
		Labels: map[Kind]string{ActEnd: "Done"}}}
	items := shownItems(b, false)
	want := []string{"Move", "View", "Aim", "Turn", "Quick", "Done"}
	if len(items) != len(want) {
		t.Fatalf("items %v", items)
	}
	for i, w := range want {
		if items[i].Text != w {
			t.Errorf("item %d: %q, want %q", i, items[i].Text, w)
		}
	}
	hk := func(its ...barItem) (ks []uint16) {
		for _, h := range hotkeysOf(its) {
			ks = append(ks, h.key)
		}
		return ks
	}
	if ks := hk(barItem{Text: "Done"}, barItem{Text: "Done", Hot: 'x'}, barItem{Text: "Move", Hot: -1}, barItem{Text: "Move"}); ks[0] != 0 || ks[1] != 'x' || ks[2] != 0 || ks[3] != 'm' {
		t.Errorf("hotkeys %v", ks)
	}
}

// TestHotkeysUnique: Use and Quick both want U; the table's Quick keeps it, Use has none.
func TestHotkeysUnique(t *testing.T) {
	hs := hotkeysOf([]barItem{{'U', "Use", ActItems, 0}, {'Q', "Quick", ActQuick, 0}, {'V', "Vow", ActWait, 0}})
	if hs[0].key != 0 || hs[1].key != 'u' || hs[1].label != "U" || hs[2].key != 'v' {
		t.Errorf("hotkeys %+v", hs)
	}
}
