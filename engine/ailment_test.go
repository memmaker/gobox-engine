package engine

import (
	"slices"
	"testing"

	"gobox/rules1e"
)

// TestAilmentShown: a weakened member (affect 0x2B, the disease's Strength
// drain) says so on the party's card (Ailment) and in the sheet's effects.
func TestAilmentShown(t *testing.T) {
	const weakened = rules1e.AffectType(0x2B)
	e := quickStarted(t)
	c := e.state().Party[0]
	if m := commonOf(e.build()).Party[0]; m.Ailment != "" {
		t.Fatalf("ailing before: %q", m.Ailment)
	}
	e.state().AddAffect(c, weakened, 60, 0, 1)
	e.rebuild()
	if m := commonOf(e.build()).Party[0]; m.Ailment != "Weakened" {
		t.Errorf("Ailment %q, want Weakened", m.Ailment)
	}
	if fx := e.state().AffectList(c); !slices.Contains(fx, "Weakened") {
		t.Errorf("effects %v lack Weakened", fx)
	}
}
