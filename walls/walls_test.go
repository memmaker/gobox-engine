package walls

import "testing"

func TestPool(t *testing.T) {
	seen := map[string]bool{}
	for _, n := range Names() {
		if seen[n] {
			t.Errorf("duplicate name %s", n)
		}
		seen[n] = true
		if Image(n) == nil {
			t.Errorf("%s does not decode", n)
		}
	}
	if len(seen) == 0 || Image("nope") != nil || len(Textures("floor_water", "nope")) != 1 {
		t.Error("pool lookups wrong")
	}
}
