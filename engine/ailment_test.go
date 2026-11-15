package engine

import (
	"slices"
	"testing"

	"gobox/rules1e"
)

// The affects of the disease's Strength drain.
const (
	affHelpless = rules1e.AffectType(0x1F)
	affWeakened = rules1e.AffectType(0x2B)
)

// TestAilmentShown: a member up but ailing has its most drastic affect for
// status, not "Okay": Weakened, then Helpless over it; the sheet's effects
// list both.
func TestAilmentShown(t *testing.T) {
	e := quickStarted(t)
	c := e.state().Party[0]
	status := func() string { e.rebuild(); return commonOf(e.build()).Party[0].Status }
	if s := status(); s != "Okay" {
		t.Fatalf("status before: %q", s)
	}
	e.state().AddAffect(c, affWeakened, 60, 0, 1)
	if s := status(); s != "Weakened" {
		t.Errorf("status %q, want Weakened", s)
	}
	e.state().AddAffect(c, affHelpless, 0, 0xFF, 0)
	if s := status(); s != "Helpless" {
		t.Errorf("status %q, want Helpless (the more drastic)", s)
	}
	if fx := e.state().AffectList(c); !slices.Contains(fx, "Weakened") || !slices.Contains(fx, "Helpless") {
		t.Errorf("effects %v", fx)
	}
}

// TestWeaknessOnceAndCured (Felix's rulings): no affect stacks, so a second
// weakening leaves one Weakened; it drains Strength as it renews itself, and
// its cure gives every point back.
func TestWeaknessOnceAndCured(t *testing.T) {
	e := quickStarted(t)
	s := e.state()
	c := s.Party[0]
	was := c.Stats[rules1e.StatStr]
	count := func() int {
		return len(slices.DeleteFunc(slices.Clone(c.Affects), func(a *rules1e.Affect) bool { return a.Type != affWeakened }))
	}
	s.CallAffect(affWeakened, c, &rules1e.Affect{}, 0) // a weakening touch
	s.CallAffect(affWeakened, c, &rules1e.Affect{}, 0) // and another
	if count() != 1 {
		t.Fatalf("%d Weakened affects after two touches, want 1", count())
	}
	if got := c.Stats[rules1e.StatStr]; got != was-2 {
		t.Fatalf("Strength %d after two touches, want %d", got, was-2)
	}
	s.RemoveAffect(c, affWeakened, nil) // its time ran out: it renews itself and drains again
	if count() != 1 || c.Stats[rules1e.StatStr] != was-3 {
		t.Fatalf("after a renewal: %d Weakened, Strength %d, want 1 and %d", count(), c.Stats[rules1e.StatStr], was-3)
	}
	s.Rv.Cure = 1
	cured := s.CureAffect(c, affWeakened)
	s.Rv.Cure = 0
	if !cured || count() != 0 || c.Stats[rules1e.StatStr] != was {
		t.Errorf("cured %v: %d Weakened, Strength %d, want none and %d", cured, count(), c.Stats[rules1e.StatStr], was)
	}
}
