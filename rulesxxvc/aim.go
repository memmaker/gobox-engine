package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// aimRefusal is ManualAimCursor's word for a square it does not accept: the menu offers "Exit" in place of "Attack".
// The source has no other reason text (TargetWhyFact is explicitly none).
const aimRefusal = "Exit"

// combatOptionAllowed is combat.pas CombatOptionAllowed: not sprinting, and a Slug weapon with no shots left is out.
func (r *Rules) combatOptionAllowed(c *Character) bool {
	if c.Status == StSprinting {
		return false
	}
	if ci, _, ty, ok := r.readied(c, xxvc.SlotWeapon); ok && ty.WeaponClass == xxvc.Slug && ci.Ref.Count <= 0 {
		return false
	}
	return true
}

// aimAccepts is ManualAimCursor's isAcceptable for an attack aimed from the actor's cell aat at the cell at: a WalkLine
// reaches it within rng (distance div 2), it is not impassable, CombatOptionAllowed, and either a loaded area weapon
// (any such cell) or a bare square with ground allowed / a fighter there who is an enemy (CanTarget) and not the actor
// (a corpse's square is no target), a ranged weapon having shots left (FindDamagingRangedWeapon). It answers who stands there.
func (r *Rules) aimAccepts(b rules.Board, c *Character, aat, at rules.BoardPos, rng int, ground bool) (who rules.CharID, ok bool) {
	who, taken := r.occupant(b, at)
	_, dist, reached, _ := walkLine(b, aat, at, 255, true, false)
	if !reached || dist/2 > rng || terrainRow(terrainAt(b, at)).cost == impassable || !r.combatOptionAllowed(c) {
		return who, false
	}
	ci, _, ty, armed := r.readied(c, xxvc.SlotWeapon)
	if r.hasAreaWeapon(c) && ci.Ref.Count > 0 { // HasLoadedAreaWeapon
		return who, true
	}
	if !taken && !ground {
		return who, false
	}
	if t := r.chars[who]; taken && (t == nil || t.team() == c.team()) { // CanTarget: the other side
		return who, false
	}
	if armed && ty.Range > 1 && (ci.Ref.Count <= 0 || ty.WeaponClass <= xxvc.Melee) { // HasRangedWeapon without FindDamagingRangedWeapon
		return who, false
	}
	return who, true
}

// aimOffer is cbtacts.pas CombatAimMenu's opening: BuildAimTargets (BuildTargetList round the actor, every enabled
// fighter a line reaches, nearest first) kept for CanTarget (the other side); Range is the weapon's (255 asks the
// readied weapon's, 0 is 1). Hits is the chance in percent where ShowRange and the foe is accepted, else -1.
// The Launcher/Missiles alternative (f.Ranged) is AD&D's: XXVc has one weapon slot.
func (r *Rules) aimOffer(id rules.CharID, f rules.AimFact) rules.AimOffer {
	c := r.chars[id]
	b := r.board(rules.Board{})
	rng := f.Range
	if rng == 0xFF {
		rng = r.readiedWeaponRange(c)
	}
	rng = max(rng, 1)
	out := rules.AimOffer{Range: rng}
	aat, ok := unitAt(rules.Fight{Board: b}, id)
	if !ok {
		return out
	}
	i := slices.IndexFunc(b.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	for _, t := range r.lineTargets(b, cellsOf(b.Units[i]), 255, true, func(e *Character) bool { return true }) {
		if t.id == id {
			continue
		}
		out.Near = append(out.Near, t.id)
		d := r.chars[t.id]
		if d.team() == c.team() {
			continue
		}
		out.Targets = append(out.Targets, t.id)
		hit := -1
		if tat, _ := unitAt(rules.Fight{Board: b}, t.id); f.ShowRange != 0 {
			if _, ok := r.aimAccepts(b, c, aat, tat, rng, false); ok {
				v, _, _ := r.hitFact(id, t.id)
				hit = int(v.(rules.Number))
			}
		}
		out.Hits = append(out.Hits, hit)
	}
	return out
}

// aimCheck is ManualAimCursor's verdict on the cell At: who stands there, "" when it is acceptable (aimAccepts),
// else aimRefusal.
func (r *Rules) aimCheck(id rules.CharID, f rules.AimCheckFact) rules.AimCheck {
	c := r.chars[id]
	b := r.board(rules.Board{})
	aat, _ := unitAt(rules.Fight{Board: b}, id)
	who, ok := r.aimAccepts(b, c, aat, f.At, f.Range, f.Ground)
	_, taken := r.occupant(b, f.At)
	out := rules.AimCheck{Target: who, Occupied: taken}
	if !ok {
		out.Why = aimRefusal
	}
	return out
}
