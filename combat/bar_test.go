package combat

import (
	"image"
	"testing"
)

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

// TestHotkeysUnique: a button takes its action's table letter whatever the
// game labels it (Use is Items: I, Vow is Wait: P); Quick keeps U.
func TestHotkeysUnique(t *testing.T) {
	hs := hotkeysOf([]barItem{{'U', "Use", ActItems, 0, 0, ""}, {'Q', "Quick", ActQuick, 0, 0, ""}, {'V', "Vow", ActWait, 0, 0, ""}})
	if hs[0].key != 'i' || hs[1].key != 'u' || hs[1].label != "U" || hs[2].key != 'p' {
		t.Errorf("hotkeys %+v", hs)
	}
}

// The game's letters only: Attack keyed A by the game is not also F (COTAB's
// arena, where F is no button: Flee is not allowed).
func TestTableKeyOnlyUnkeyed(t *testing.T) {
	b := bar{items: []barItem{{Text: "Attack", Kind: ActAttack, Hot: 'A'}, {Text: "Quick", Kind: ActQuick}}, rects: make([]image.Rectangle, 2)}
	if it := b.tableKey('f'); it != nil {
		t.Errorf("F pressed %q", it.Text)
	}
	if it := b.tableKey('u'); it == nil || it.Text != "Quick" {
		t.Errorf("U on an unkeyed Quick: %v", it)
	}
}
