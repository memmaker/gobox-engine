package sprites

import "testing"

func TestPools(t *testing.T) {
	for _, p := range []string{"portraits", "combat", "items"} {
		seen := map[string]bool{}
		for _, n := range Names(p) {
			if seen[n] {
				t.Errorf("%s: duplicate %s", p, n)
			}
			seen[n] = true
			if Image(p, n) == nil {
				t.Errorf("%s/%s does not decode", p, n)
			}
		}
	}
	if len(Names("items")) == 0 || Image("items", "nope") != nil {
		t.Error("pool lookups wrong")
	}
}
