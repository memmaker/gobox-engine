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
	if h := hotkeyOf("Done"); h.key != 'd' || h.label != "D" {
		t.Errorf("Done's hotkey %+v", h)
	}
	if h := hotkeyOf("Move"); h.key != 'm' {
		t.Errorf("Move's hotkey %+v", h)
	}
}
