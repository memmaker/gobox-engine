package rules1e

import "gobox/rules"

// A fight's reserve and its reinforcing (extra to POR; pod-re
// CharFightsAlone and its Monster Summoning, generic here: any adventure's
// table fills a reserve, any effect may reinforce from it).
//
// A reserve unit is a monster added for the fight with StatusReserve: it
// keeps a combatant's slot but stands on no square, takes no turn and
// counts for nothing (the monsters' health, the fight's end, experience,
// treasure) until reinforce brings it into play.

// summonedMost is how many a Monster Summoning gates in (pod-re magic.pas: placed = 2).
const summonedMost = 2

// holdInReserve resolves rules.HoldInReserve: before the fight is set up.
func (r *Rules) holdInReserve(a rules.HoldInReserve) error {
	c := r.chars[a.Who]
	if c == nil {
		return errNoChar
	}
	if c.Side != SideMonsters || r.S.BeginFlag == 5 {
		return ErrNotNow
	}
	c.Status = StatusReserve
	return nil
}

// bareGround is the fight's plain ground, the only one a reinforcement
// arrives on: a dungeon's floor, outdoors the ground the map is filled with.
func (e *RuleState) bareGround() Ground {
	if w := e.T().Gen.Wild; !e.Dungeon && w != nil {
		return w.Grass
	}
	return e.grounds().Floor
}

// Reinforce = pod-re magic.pas:2870 (Monster Summoning): up to most units
// of the reserve, in its order, join by's side, each on the first free
// square of bare ground beside a combatant: the combatants are tried in
// the fight's order up to the reserve, by itself last, and each one's
// eight squares in turn. It tells how many came ("summons is not
// answered" for none).
//
// As the original: the squares already tried beside a combatant are not
// tried again for the next unit, and the unit about to come has changed
// sides even when no square is found for it. Fixed: a square off the map
// is passed over (the original never left it: Inc(dir) is inside the
// bounds check).
func (e *RuleState) Reinforce(by *Character, most int) int {
	var reserve []*Character
	for _, c := range e.Party {
		if c.Status == StatusReserve {
			reserve = append(reserve, c)
		}
	}
	bare := e.bareGround()
	placed, cand, dir, done := 0, 0, 0, false
	for !done {
		if dir > 7 {
			if cand+1 >= len(e.Party) || e.Party[cand].Status == StatusReserve {
				cand, done = -1, true // by itself, last
			} else {
				cand++
			}
			dir = 0
		}
		if len(reserve) == 0 || placed == most {
			break
		}
		unit, at := reserve[0], by
		if cand >= 0 {
			at = e.Party[cand]
		}
		unit.Side = by.Side
		x, y := e.CombatX(at), e.CombatY(at)
		for found := false; !found && dir <= 7; dir++ {
			nx, ny := x+DirDX[dir], y+DirDY[dir]
			if who, g := e.MapAt(nx, ny); who != 0 || g != bare {
				continue
			}
			if found = e.reviveAt(unit, nx, ny, unit.MaxHP, "gates in"); found {
				placed++
				reserve = reserve[1:]
				done = done || len(reserve) == 0
			}
		}
	}
	if placed == 0 {
		e.CombatSay(by, "summons is not answered", true)
	}
	return placed
}
