package sprites

import "testing"

func TestPools(t *testing.T) {
	for _, p := range []string{"portraits", "combat", "items"} {
		seen := map[string]bool{}
		for _, n := range SpriteNames("pool", p) {
			if seen[n] {
				t.Errorf("%s: duplicate %s", p, n)
			}
			seen[n] = true
			if SpriteImage("pool", p, n) == nil {
				t.Errorf("%s/%s does not decode", p, n)
			}
		}
	}
	if len(SpriteNames("pool", "items")) == 0 || SpriteImage("pool", "items", "nope") != nil || len(SpriteNames("buck-rogers", "items")) != 0 {
		t.Error("pool lookups wrong")
	}
}
