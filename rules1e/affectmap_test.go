package rules1e

import (
	"maps"
	"testing"
)

// TestAffectMapRoundTrip: game numbers become ids and come back, items'
// granted affects included; an unmapped number is an error.
func TestAffectMapRoundTrip(t *testing.T) {
	tb := POR
	tb.AffectIDs = AffectMap{3: 0x90, 1: AffBless, 0x80: AffItem0, 7: affStrength}
	s := RuleState{Tables: &tb, Items: make([]ItemBase, 1)}
	a := []AffectRecord{{Type: 3}, {Type: 1}}
	it := []Item{{}}
	it[0].Spells[2], it[0].Spells[1] = 0x80, 7
	if err := s.toIDs(a, it); err != nil {
		t.Fatal(err)
	}
	if a[0].Type != 0x90 || a[1].Type != AffBless || it[0].Spells[1] != ItemPower(affStrength) {
		t.Fatalf("ids %v %v", a, it[0].Spells)
	}
	if err := s.toGame(a, it); err != nil {
		t.Fatal(err)
	}
	if a[0].Type != 3 || a[1].Type != 1 || it[0].Spells != [3]ItemPower{0, 7, 0x80} {
		t.Fatalf("back %v %v", a, it[0].Spells)
	}
	if err := s.toIDs([]AffectRecord{{Type: 9}}, nil); err == nil {
		t.Fatal("unmapped number passed")
	}
}

// TestGameAffectIDs: every number of each game's map comes back from its id
// (one number per id, so Export is exact).
func TestGameAffectIDs(t *testing.T) {
	for name, m := range map[string]AffectMap{"cotab": CotabAffectIDs, "sotsb": SotsbAffectIDs, "pod": PodAffectIDs} {
		tb := Tables{AffectIDs: m}
		for b := range m {
			id, err := tb.affectID(b)
			if err != nil {
				t.Fatal(err)
			}
			if back, err := tb.gameAffect(id); err != nil || back != b {
				t.Errorf("%s: %d -> %#x -> %d (%v)", name, b, id, back, err)
			}
		}
	}
}

// TestAffectHandlersFull (BUG-CLASSES 25): ovr13:31A2 fills POR's slots
// 1..0x8B (0x58, breath, by the spell setup), and every id of a game's
// affectIds has a handler in that game's table (POR's with its
// affectVariants): callAffect never skips a mapped id.
func TestAffectHandlersFull(t *testing.T) {
	for i := AffectType(1); i < 0x8C; i++ {
		if affectHandler[POR.Affects[i]] == nil {
			t.Errorf("POR affect %#x has no handler", i)
		}
	}
	for name, g := range map[string]struct {
		ids      AffectMap
		powers   PowerMap
		variants map[AffectType]HandlerName
	}{"cotab": {CotabAffectIDs, CotabItemPowers, CotabAffectVariants}, "sotsb": {SotsbAffectIDs, SotsbItemPowers, SotsbAffectVariants}, "pod": {PodAffectIDs, PodItemPowers, PodAffectVariants}} {
		unported := UnportedAffectHandlers[name]
		a := maps.Clone(POR.Affects)
		for id, h := range g.variants {
			if _, listed := unported[h]; listed {
				if affectHandler[h] != nil {
					t.Errorf("%s: %q has a handler now: take it off UnportedAffectHandlers", name, h)
				}
				continue
			}
			if affectHandler[h] == nil {
				t.Errorf("%s: affectVariants %#x names no handler %q", name, id, h)
			}
			a[id] = h
		}
		skip := func(id AffectType) bool { _, listed := unported[g.variants[id]]; return listed }
		for b, id := range g.ids {
			if !skip(id) && affectHandler[a[id]] == nil {
				t.Errorf("%s: affect %d (id %#x) has no handler", name, b, id)
			}
		}
		for b, p := range g.powers {
			if !skip(AffectType(p)) && affectHandler[a[AffectType(p)]] == nil {
				t.Errorf("%s: item power %d (%#x) has no handler", name, b, p)
			}
		}
	}
}

// TestGameItemPowers: each game's item power codes come back from their
// unified powers, and an item's granted affect goes through the affect map
// only for a granting power.
func TestGameItemPowers(t *testing.T) {
	for name, m := range map[string]PowerMap{"cotab": CotabItemPowers, "sotsb": SotsbItemPowers, "pod": PodItemPowers} {
		tb := Tables{PowerIDs: m}
		for b := range m {
			p, err := tb.powerID(b)
			if err != nil {
				t.Fatal(err)
			}
			if back, err := tb.gamePower(p); err != nil || back != b {
				t.Errorf("%s: %d -> %#x -> %d (%v)", name, b, p, back, err)
			}
		}
	}
	tb := POR
	tb.AffectIDs, tb.PowerIDs = AffectMap{38: affStrength}, CotabItemPowers
	s := RuleState{Tables: &tb, Items: make([]ItemBase, 1)}
	it := []Item{{}, {}}
	it[0].Spells[2], it[0].Spells[1] = 0x80, 38 // grants affect 38
	it[1].Spells[2], it[1].Spells[1] = 0x85, 3  // giant strength 21: a value
	if err := s.toIDs(nil, it); err != nil {
		t.Fatal(err)
	}
	if it[0].Spells != [3]ItemPower{0, ItemPower(affStrength), ItemPower(AffItem0)} || it[1].Spells != [3]ItemPower{0, 3, ItemPower(itemPowerGiantStrength)} {
		t.Fatalf("ids %v %v", it[0].Spells, it[1].Spells)
	}
	if err := s.toGame(nil, it); err != nil {
		t.Fatal(err)
	}
	if it[0].Spells != [3]ItemPower{0, 38, 0x80} || it[1].Spells != [3]ItemPower{0, 3, 0x85} {
		t.Fatalf("back %v %v", it[0].Spells, it[1].Spells)
	}
}
