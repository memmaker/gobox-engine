package rules1e

import "gobox/rules"

// The combat map's size in squares.
const (
	combatMapWidth  = 50
	combatMapHeight = 25
)

// coneRange: the cone's sight line is not cut short by range.
const coneRange = 0x7FFF

// coneFoundMax is the most combatants a cone collects (cotab's 30-entry list).
const coneFoundMax = 30

// The squares beside a cone's end, by the direction of its last step.
var (
	coneSideAX = [8]int{-1, 0, 0, 1, 1, 0, 0, -1}
	coneSideAY = [8]int{0, -1, -1, 0, 0, 1, 1, 0}
	coneSideBX = [8]int{1, 1, 0, 0, -1, -1, 0, 0}
	coneSideBY = [8]int{0, 0, 1, 1, 0, 0, -1, -1}
)

// clampToMap = cotab O0C2:16B7 (var x, var y): onto the combat map.
func clampToMap(x, y *int) {
	*x = min(max(*x, 0), combatMapWidth-1)
	*y = min(max(*y, 0), combatMapHeight-1)
}

// coneTargets = cotab O0C2:17CF (x0, y0, x1, y1, width, length): the
// combatants in a cone from (x0,y0) towards (x1,y1): the line is drawn on,
// repeating its steps, to length squares (a diagonal step counts 1.5),
// stopped by walls, and walked; widths 2 and 3 add the lines to the squares
// beside its end. Everyone found but the caster becomes a target.
func (e *RuleState) coneTargets(x0, y0, x1, y1, width, length int) {
	l := rules.NewLine(x0, y0, x1, y1)
	var dirs []Dir
	for l.Step() {
		dirs = append(dirs, Dir(l.Dir))
	}
	last := len(dirs) - 1 // a byte in the asm: 255 with no step
	i := 0
	length *= 2
	walked := byte(l.Steps)
	for walked < byte(length) {
		if x1 >= combatMapWidth-1 || x1 <= 0 || y1 >= combatMapHeight-1 || y1 <= 0 {
			break
		}
		d := Dir(0)
		if i < len(dirs) {
			d = dirs[i]
		}
		if d%2 == 0 {
			walked += 2
		} else {
			walked += 3
		}
		x1 += DirDX[d]
		y1 += DirDY[d]
		if i == last {
			i = 0
		} else {
			i++
		}
	}
	clampToMap(&x1, &y1)
	rng := coneRange
	e.LineOfSight(&rng, &x1, &y1, x0, y0)
	var found [coneFoundMax + 1]Slot // [1..]: the combatants met, 0 ends the list
	n := 1
	dup := false
	var lastDir Dir
	l = rules.NewLine(x0, y0, x1, y1)
	for l.Step() {
		who, _ := e.MapAt(l.X, l.Y)
		if who > 0 {
			if n > 1 {
				for k := 1; k <= n-1; k++ {
					if found[k] == who {
						dup = true // never cleared again: later combatants are left out
					}
				}
			} else {
				dup = false
			}
			if !dup && n < len(found) {
				found[n] = who
				n++
			}
			lastDir = Dir(l.Dir) // only on a square with a combatant (asm skips it otherwise)
		}
	}
	side := func(sx, sy int) {
		s := rules.NewLine(x0, y0, sx, sy)
		for s.Step() {
			who, _ := e.MapAt(s.X, s.Y)
			if who == 0 {
				continue
			}
			for k := 1; k < len(found); k++ {
				if found[k] == who {
					break
				}
				if found[k] == 0 {
					found[k] = who
					break
				}
			}
		}
	}
	if width > 1 {
		ax, ay := x1+coneSideAX[lastDir%8], y1+coneSideAY[lastDir%8]
		clampToMap(&ax, &ay)
		bx, by := x1+coneSideBX[lastDir%8], y1+coneSideBY[lastDir%8]
		clampToMap(&bx, &by)
		side(bx, by)
		if width > 2 {
			side(ax, ay)
		}
	}
	e.NSpellTargets = 0
	sel := e.Acting()
	for k := 1; k < len(found) && found[k] != 0; k++ {
		if c := e.Combatant(byte(found[k])); c != sel {
			e.NSpellTargets++
			e.SetTgt(e.NSpellTargets, c)
		}
	}
}

// Spell 98's cone: 1 wide, 3 long, 6d6.
const (
	spell98ConeWidth   = 1
	spell98ConeLength  = 3
	spell98Dice        = 6
	spell98Sides       = 6
	kindSpell98Saves   = 18 // the only kind that rolls a save against it
	spellSaveModeField = 8
)

// fxSpell98 = cotab O0C2:5727: 6d6 in a narrow cone, taken as saved
// (halved) except by creatures of kind 18.
func (e *RuleState) fxSpell98() {
	sel := e.Acting()
	e.coneTargets(e.CombatX(sel), e.CombatY(sel), e.TgtX, e.TgtY, spell98ConeWidth, spell98ConeLength)
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t == nil {
			continue
		}
		e.Damage(t, e.RollDice(spell98Dice, spell98Sides), e.SpellRec(spUnnamed98, spellSaveModeField), t.Kind != kindSpell98Saves)
	}
}
