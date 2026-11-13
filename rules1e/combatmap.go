package rules1e

import (
	"gobox/rules"
	"slices"
)

// Combat map services (ovr31/ovr32 and the START sprite blitter) that the
// rules call: the near list and the helpers over the combat map.

// NearEntry is one 3-byte record of the DS:6718 list built by ovr31:0909.
type NearEntry struct{ Index, dist, dir byte }

// Combatant = DS:65B9[i]: the Combatant behind a near list index.
func (e *RuleState) Combatant(i byte) *Character {
	if int(i) < 1 || int(i) > len(e.CombatantSlots) {
		return nil
	}
	return e.CombatantSlots[i-1]
}

// shoot = flyMissile (00BA:1A29): something flies from square (x1, y1)
// to (x2, y2), a missile item or a spell or power; present draws it (Shot).
func (e *RuleState) shoot(x1, y1, x2, y2 int, item rules.ItemID, spell rules.SpellID) {
	if !e.Fight.On || x1 == x2 && y1 == y2 { // the original's line of fewer than two steps: nothing flies
		return
	}
	who, _ := e.MapAt(x1, y1) // the shooter stands where it flies from
	e.Emit(rules.Shot{From: unit(int(who)), Start: rules.BoardPos{X: x1, Y: y1}, To: rules.BoardPos{X: x2, Y: y2}, Item: item, Spell: spell})
}

// IsParty = 00BA:23B7 (char): true for the party side.
func IsParty(c *Character) bool { return c.Side == SideParty }

// NearEnemies = 00BA:2433 (char, range): combatants of the other side
// within range; the near list keeps only them, copied to DS:6DE4.
func (e *RuleState) NearEnemies(c *Character, r int) int {
	e.nearCombatants(c, e.CombatX(c), e.CombatY(c), r)
	var want byte // the original keeps the targets whose side byte equals isParty(c)
	if IsParty(c) {
		want = 1
	}
	kept := slices.DeleteFunc(e.Near, func(n NearEntry) bool {
		t := e.Combatant(n.Index)
		return t == nil || t.Side != want
	})
	e.Near = kept
	for i, n := range kept {
		e.Targets[i+1] = Slot(n.Index)
	}
	return len(kept)
}

// RangeTo = 00BA:255B (char, target): distance in squares (list distance
// halved) from char to target, the near list left as it was.
func (e *RuleState) RangeTo(c, t *Character) int {
	saved := e.Near
	e.Near = append([]NearEntry(nil), saved...)
	e.Cmap.seeAll = 1 // [DS:6716]+6: line of sight ignores walls
	e.nearCombatants(c, e.CombatX(c), e.CombatY(c), 0xFF)
	e.Cmap.seeAll = 0
	var d byte
	for i, n := range e.Near {
		if e.Combatant(n.Index) == t || i == len(e.Near)-1 {
			d = n.dist
			break
		}
	}
	// only the 0xD8 entry bytes come back, the count (DS:671A) stays
	list := make([]NearEntry, len(e.Near))
	copy(list, saved)
	e.Near = list
	return int(d >> 1)
}
