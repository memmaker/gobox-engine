package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
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
	status := func() (string, rules.Health) {
		e.rebuild()
		m := commonOf(e.build()).Party[0]
		return m.Status, m.Health
	}
	if s, h := status(); s != "Okay" || h != rules.Hale {
		t.Fatalf("before: %q, health %d", s, h)
	}
	e.state().AddAffect(c, affWeakened, 60, 0, 1)
	if s, h := status(); s != "Weakened" || h != rules.Ailing {
		t.Errorf("%q, health %d, want Weakened and ailing", s, h)
	}
	e.state().AddAffect(c, affHelpless, 0, 0xFF, 0)
	if s, h := status(); s != "Helpless" || h != rules.Disabled { // the more drastic; it takes the turns
		t.Errorf("%q, health %d, want Helpless and disabled", s, h)
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

// TestTempleCuresWeakness: the temple's Cure Disease is offered for a
// weakened, helpless member (by its index, as gobox sends it), takes both
// affects off and gives the Strength back.
func TestTempleCuresWeakness(t *testing.T) {
	e := devPlay(t, "temple")
	s := e.state()
	c := s.Party[0]
	was := c.Stats[rules1e.StatStr]
	s.CallAffect(affWeakened, c, &rules1e.Affect{}, 0)
	s.AddAffect(c, affHelpless, 0, 0xFF, 0)
	s.Purse.Gold = 10000
	e.rebuild()
	tr := e.build().(game.Trading)
	disease := slices.IndexFunc(s.T().Town.Temple, func(sv rules1e.TempleService) bool { return sv.Kind == rules1e.TempleDisease })
	if disease < 0 || !slices.Contains(tr.Services[disease].Needs, 0) {
		t.Fatalf("no disease cure offered for the member: %+v", tr.Services)
	}
	if _, err := e.doNow(game.Heal{Member: 0, Service: disease}); err != nil {
		t.Fatal("Heal:", err)
	}
	if s.FindAffect(c, affWeakened) != nil || s.FindAffect(c, affHelpless) != nil || c.Stats[rules1e.StatStr] != was {
		t.Errorf("after the cure: weakened %v, helpless %v, Strength %d (was %d)", s.FindAffect(c, affWeakened) != nil, s.FindAffect(c, affHelpless) != nil, c.Stats[rules1e.StatStr], was)
	}
}
