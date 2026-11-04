package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"por/rules"
)

// sameFlags says the flags classic takes from State equal por's: the
// view with them blanked, then filled by classic, is por's.
func sameFlags(t *testing.T, name string, e *rules.Engine, seen map[string]int) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	b := v
	b.Exploring, b.Sweeping, b.Interact, b.CanCast, b.Open = false, false, false, false, [4]bool{}
	b.Adventure, b.Day, b.Inside, b.Ordering, b.Credits, b.Full, b.IconDrop = false, false, false, false, false, false, false
	c := classic.Flags(b, s)
	got := []any{c.Exploring, c.Sweeping, c.Interact, c.CanCast, c.Open, c.Adventure, c.Day, c.Inside, c.Ordering, c.Credits, c.Full, c.IconDrop}
	want := []any{v.Exploring, v.Sweeping, v.Interact, v.CanCast, v.Open, v.Adventure, v.Day, v.Inside, v.Ordering, v.Credits, v.Full, v.IconDrop}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: flags\n got %v\nwant %v", name, got, want)
	}
	for k, on := range map[string]bool{"exploring": v.Exploring, "adventure": v.Adventure, "day": v.Day, "inside": v.Inside, "full": v.Full, "open": v.Open != [4]bool{}} {
		if on {
			seen[k]++
		}
	}
}

// TestFlagsParity: the status flags at each stop of the dev scenes, Enter on.
func TestFlagsParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "password", "parlay", "who", "journal", "runes", "lockeddoor", "shop", "temple", "loot", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 12; i++ {
			sameFlags(t, scene, e, seen)
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	for _, k := range []string{"exploring", "adventure", "open"} {
		if seen[k] == 0 {
			t.Errorf("no %s seen: %v", k, seen)
		}
	}
	t.Logf("flags compared: %v", seen)
}
