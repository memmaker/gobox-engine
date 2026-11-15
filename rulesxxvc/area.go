package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// areaReach is combatai.pas AREA_REACH: the farthest SelectAreaTarget scores
// cells from its origin.
const areaReach = 9

// areaHalfWidths is gdconsts.pas areaHalfWidths[-3..3, 1..3]: a blast's half
// width by row offset and area class; negative rows are outside it.
var areaHalfWidths = [7][3]int{
	{-3, -3, 1}, {-3, 1, 2}, {1, 2, 3}, {1, 2, 3}, {1, 2, 3}, {-3, 1, 2}, {-3, -3, 1},
}

// areaTarget is SelectAreaTarget's choice: the best cell and its weight.
type areaTarget struct {
	at     rules.BoardPos
	weight int
}

// selectAreaTarget is combatai.pas SelectAreaTarget's scoring with weapon
// (ci, tt, ty) as id's readied weapon: every enabled combatant within the
// outer range adds its side's weight to the cells of a blast that would
// cover it, and the best cell a line (stopping at chaff) reaches from it
// to id is chosen, a tie taken on a 25% roll. ok is false for a weapon not
// of an area class.
func (r *Rules) selectAreaTarget(f rules.Fight, id rules.CharID, c *Character, s *combatState, tt xxvc.ItemTemplate, ty xxvc.ItemType, rng *rules.RNG) (areaTarget, bool) {
	var class int
	switch ty.WeaponClass {
	case xxvc.RocketLauncher:
		class = 2
	case xxvc.Grenade:
		class = 1
	case xxvc.Plasma:
		class = 3
	default:
		return areaTarget{}, false
	}
	cloud := tt.Type == chaffGrenadeType || tt.Type == aerosolMistGrenadeType
	weaponRange := r.weaponRange(c, ty, true)
	outer := weaponRange + class
	originID := id
	if outer > areaReach {
		outer = areaReach
		if !cloud {
			originID = s.target
		}
	}
	inner := outer - class
	var weights [2]int8 // C_Side.TEAM, C_Side.HOSTILE
	switch {
	case cloud && c.team():
		weights = [2]int8{4, 1}
	case cloud:
		weights = [2]int8{1, 4}
	case c.team():
		weights = [2]int8{-12, 1}
	default:
		weights = [2]int8{1, -2}
	}
	origin, ok := unitAt(f, originID)
	if !ok {
		return areaTarget{}, true
	}
	var scores [2*areaReach + 1][2*areaReach + 1]int8 // [x offset][y offset]
	for _, uid := range r.fighters(f) {
		u, p := r.chars[uid], rules.BoardPos{}
		if q, ok := unitAt(f, uid); ok {
			p = q
		} else {
			continue
		}
		if !u.Enabled || p.X < origin.X-outer || p.X > origin.X+outer || p.Y < origin.Y-outer || p.Y > origin.Y+outer {
			continue
		}
		w := weights[side(u)]
		switch tt.Type {
		case chaffGrenadeType, aerosolMistGrenadeType:
			if t := terrainAt(f.Board, p); t == chaffTerrain || t == aerosolMistTerrain {
				w = -1
			}
		case stunGrenadeType:
			if hasEffect(u, EffStunned) || helpless(u) {
				w = 0
			}
		case gasGrenadeType:
			if helpless(u) {
				w = 0
			}
		}
		lo, hi := max(-class, origin.Y-inner-p.Y), min(class, origin.Y+inner-p.Y)
		for i := lo; i <= hi; i++ {
			half := areaHalfWidths[i+3][class-1]
			for x := max(p.X-half, origin.X-inner); x <= min(p.X+half, origin.X+inner); x++ {
				scores[x-origin.X+areaReach][i+p.Y-origin.Y+areaReach] += w
			}
		}
	}
	at, _ := unitAt(f, id)
	best := areaTarget{}
	var bestDX, bestDY int
	for j := -inner; j <= inner; j++ {
		for i := -inner; i <= inner; i++ {
			w := int(scores[j+areaReach][i+areaReach])
			if w > best.weight || w == best.weight && w > 0 && rng.N(100) < 25 {
				cell := rules.BoardPos{X: origin.X + j, Y: origin.Y + i}
				if !onBoard(f.Board, cell) {
					continue
				}
				if _, _, ok, seen := walkLine(f.Board, cell, at, weaponRange, true, true); ok && !seen.chaff {
					bestDX, bestDY, best.weight = j, i, w
				}
			}
		}
	}
	best.at = rules.BoardPos{X: origin.X + bestDX, Y: origin.Y + bestDY}
	return best, true
}

// areaChoice is SelectAreaTarget mode 0's choice: the weapon's value times
// the best weight, over 128 scaled to (v shr 4) + 120 (EvaluateItem).
func areaChoice(t areaTarget, itemType int32, rng *rules.RNG) int {
	if t.weight <= 0 {
		return 0
	}
	var v int
	switch itemType {
	case chaffGrenadeType:
		v = rules.RNGDice(rng, 1, 6) + 1
	case aerosolMistGrenadeType:
		v = rules.RNGDice(rng, 1, 6)
	case explosiveGrenadeType:
		v = 40
	case gasGrenadeType:
		v = 5
	case dazzleGrenadeType:
		v = rules.RNGDice(rng, 1, 10) + 10
	case stunGrenadeType:
		v = rules.RNGDice(rng, 1, 10) + 5
	case plasmaThrowerType:
		v = 25
	case rocketLauncherType:
		v = 30
	}
	v = int(uint16(v * t.weight))
	if v > 128 {
		return uint8Wrap(v>>4 + 120)
	}
	return v
}

// side is C_Side: 0 TEAM, 1 HOSTILE.
func side(c *Character) int {
	if c.team() {
		return 0
	}
	return 1
}
