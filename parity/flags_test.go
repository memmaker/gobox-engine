package parity

import (
	"testing"

	"gobox/front"
)

// TestFlagsParity: State.Flags at each stop of the dev scenes, Enter on.
func TestFlagsParity(t *testing.T) {
	seen := map[string]int{}
	for _, scene := range []string{"", "password", "parlay", "who", "journal", "runes", "lockeddoor", "shop", "temple", "loot", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 12; i++ {
			at(t, scene, e)
			f := e.Rules().State().Flags
			if f == nil {
				t.Fatalf("%s step %d: no flags", scene, i)
			}
			for k, on := range map[string]bool{"exploring": f.Exploring, "adventure": f.Adventure, "day": f.Day, "inside": f.Inside, "full": f.Full, "open": f.Open != [4]bool{}} {
				if on {
					seen[k]++
				}
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
	for _, k := range []string{"exploring", "adventure", "open"} {
		if seen[k] == 0 {
			t.Errorf("no %s seen: %v", k, seen)
		}
	}
	t.Logf("flags seen: %v", seen)
}
