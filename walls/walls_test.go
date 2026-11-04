package walls

import "testing"

func TestPool(t *testing.T) {
	for _, s := range Series() {
		seen := map[string]bool{}
		for _, n := range Names(s) {
			if seen[n] {
				t.Errorf("%s: duplicate name %s", s, n)
			}
			seen[n] = true
			if Image(s, n) == nil {
				t.Errorf("%s/%s does not decode", s, n)
			}
		}
	}
	if len(Names("pool")) == 0 || Image("pool", "nope") != nil || Image("buck-rogers", "floor_water") != nil ||
		len(Textures("pool", "floor_water", "nope")) != 1 {
		t.Error("pool lookups wrong")
	}
}
