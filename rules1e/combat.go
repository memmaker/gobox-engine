package rules1e

// The combat main loop and the player's command menus (ovr09).

import (
	"gobox/rules"
	"slices"
	"strconv"
)

// NextCombatant = ovr09:0131 (var char): the highest initiative left (combat
// block +3); ties go to the higher d100. nil when nobody is left.
func (e *RuleState) NextCombatant() *Character {
	var best, bestRoll byte
	var who *Character
	for _, c := range e.Party {
		r := byte(e.Dice(1, 100))
		ini := c.Combat.Delay
		if ini > best {
			bestRoll = r
		}
		if ini >= best && r >= bestRoll {
			bestRoll, best, who = r, ini, c
		}
	}
	if best == 0 {
		return nil
	}
	return who
}

func B2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// FleeDir is a direction in which c's next step leaves the map (moveAttack
// offers Flee there), or -1.
func (e *RuleState) FleeDir(c *Character) int {
	for d := 0; d < 8; d++ {
		if who, g := e.GroundAt(c, d); who == 0 && g == 0 {
			return d
		}
	}
	return -1
}

// WeaponRange is the squares the readied weapon reaches (1: melee).
func (e *RuleState) WeaponRange(c *Character) int {
	if w := c.Slots[SlotWeapon]; w != nil && e.Base(w.Type).Range > 1 {
		return int(e.Base(w.Type).Range) - 1
	}
	return 1
}

// Walk (extra) takes c along path as moveAttack's keypad steps would, one
// step a message delay apart, while the next square is free. done: c is
// out (fallen, helpless).
func (e *RuleState) Walk(c *Character, path []int) (done bool) {
	ci := c.Combat
	for _, d := range path {
		if ci.Moves <= 1 || done {
			break
		}
		if who, g := e.GroundAt(c, d); who > 0 || g == 0 || byte(e.GroundInfo(Ground(g))[0]) > ci.Moves {
			break
		}
		e.TurnTo(c, byte(d))
		e.freeAttacksOnLeave(c, d)
		if c.Okay == 0 {
			return e.ClearCombatInfo(c)
		}
		e.stepSquare(c, d)
		if c.Okay == 0 {
			done = e.ClearCombatInfo(c)
		}
		e.inClouds(c)
		if e.IsHelpless(c) {
			done = e.ClearCombatInfo(c)
		}
	}
	if ci.Moves < 2 {
		ci.Moves = 0
	}
	return
}

// CombatSay (extra) is combat text for the View's Log: fresh starts a new
// message, otherwise it goes on the current one; c (if any) is named first.
func (e *RuleState) CombatSay(c *Character, msg string, fresh bool) {
	e.Emit(rules.Narrated{Text: say1(c, msg), More: !fresh})
}

func name(c *Character) string {
	if c == nil {
		return ""
	}
	return c.Name()
}

// MoveReach (extra) finds the squares c can walk to with the moves it has
// left, as moveAttack's steps allow them (2 a step, 3 diagonally, free
// ground no dearer than the moves left), each with the steps that keep
// the most moves.
func (e *RuleState) MoveReach(c *Character) map[[2]int][]int {
	idx := e.CombatIndex(c)
	ox, oy := e.Cpos[idx].X, e.Cpos[idx].Y
	defer func() { e.Cpos[idx].X, e.Cpos[idx].Y = ox, oy }()
	type node struct {
		left int
		path []int
	}
	best := map[[2]int]node{{int(ox), int(oy)}: {left: int(c.Combat.Moves)}}
	queue := [][2]int{{int(ox), int(oy)}}
	for len(queue) > 0 {
		at := queue[0]
		queue = queue[1:]
		n := best[at]
		if n.left <= 1 {
			continue
		}
		e.Cpos[idx].X, e.Cpos[idx].Y = byte(at[0]), byte(at[1])
		for d := 0; d < 8; d++ {
			cost := 2 + d&1
			who, g := e.GroundAt(c, d)
			if cost > n.left || who > 0 || g == 0 || int(e.GroundInfo(Ground(g))[0]) > n.left {
				continue
			}
			to := [2]int{at[0] + DirDX[d], at[1] + DirDY[d]}
			if b, ok := best[to]; ok && b.left >= n.left-cost {
				continue
			}
			best[to] = node{n.left - cost, append(append([]int(nil), n.path...), d)}
			queue = append(queue, to)
		}
	}
	out := map[[2]int][]int{}
	for k, n := range best {
		if len(n.path) > 0 {
			out[k] = n.path
		}
	}
	return out
}

// WalkSteps (extra) are the squares path walks c through, its own first,
// each marked when leaving it lets an enemy take a free attack.
func (e *RuleState) WalkSteps(c *Character, path []int) []Step {
	at := [2]int{e.CombatX(c), e.CombatY(c)}
	out := []Step{{At: rules.BoardPos{X: at[0], Y: at[1]}}}
	for _, d := range path {
		out[len(out)-1].Free = e.provokes(c, at, d)
		at = [2]int{at[0] + DirDX[d], at[1] + DirDY[d]}
		out = append(out, Step{At: rules.BoardPos{X: at[0], Y: at[1]}})
	}
	return out
}

// provokes (extra) is freeAttacksOnLeave's test without the attacks: does
// c stepping from square at in direction d leave an enemy that may strike
// it? Position, near list and targets are kept.
func (e *RuleState) provokes(c *Character, at [2]int, d int) bool {
	idx := e.CombatIndex(c)
	ox, oy, near, targets := e.Cpos[idx].X, e.Cpos[idx].Y, slices.Clone(e.Near), e.Targets
	defer func() { e.Cpos[idx].X, e.Cpos[idx].Y, e.Near, e.Targets = ox, oy, near, targets }()
	e.Cpos[idx].X, e.Cpos[idx].Y = byte(at[0]), byte(at[1])
	before := slices.Clone(e.Targets[1 : e.NearEnemies(c, 1)+1])
	e.Cpos[idx].X, e.Cpos[idx].Y = byte(at[0]+DirDX[d]), byte(at[1]+DirDY[d])
	after := e.Targets[1 : e.NearEnemies(c, 1)+1]
	for _, i := range before {
		t := e.Combatant(byte(i))
		if !slices.Contains(after, i) && t != nil && !e.IsHelpless(t) && e.MayActAgainst(t, c) &&
			e.FindAffect(t, affBloodDrain) == nil && e.FindAffect(t, affStrikeTied) == nil {
			return true
		}
	}
	return false
}

// RangedWeapon = 00BA:2E03 (char): the readied weapon shoots (base item
// +0xC above 1).
func (e *RuleState) RangedWeapon(c *Character) bool {
	w := c.Slots[SlotWeapon]
	return w != nil && e.Base(w.Type).Range > 1
}

// Throwable = 00BA:2E46 (char): a ranged weapon that also works in melee
// (base item flags 0x14).
func (e *RuleState) Throwable(c *Character) bool {
	return e.RangedWeapon(c) && e.Base(c.Slots[SlotWeapon].Type).Usage&0x14 == 0x14
}

// WeaponRules (extra) tell how it attacks in combat (the bar's tooltip):
// its reach, and what stops a throw or a shot (targetWhy's refusals), what
// a throw costs (fullAttack: a melee-too weapon lands on the ground, the
// others are used up).
func (e *RuleState) WeaponRules(it *Item) []string {
	b := e.Base(it.Type)
	rng := strconv.Itoa(int(b.Range) - 1)
	switch {
	case b.Range <= 1:
		return []string{"Melee: a foe next to you"}
	case b.Usage&0x14 == 0x14: // thrown, in melee too
		return []string{"Melee, or thrown up to " + rng + " squares;", "a thrown one lands on the ground"}
	case b.Usage&0x10 != 0:
		return []string{"Thrown up to " + rng + " squares,", "not with a foe next to you;", "one is used up each throw"}
	case b.Usage&8 != 0 && b.Usage&0x81 != 0: // fires arrows or bolts
		return []string{"Shoots up to " + rng + " squares", "with its ammunition readied,", "not with a foe next to you"}
	case b.Usage&8 != 0: // a sling
		return []string{"Shoots up to " + rng + " squares,", "not with a foe next to you"}
	}
	return []string{"Reach: " + rng + " squares"}
}

// AmmoFor = 00BA:2E8B (char, var item): the item a ranged attack uses up
// (the weapon itself, the arrows in slot 11 or the bolts in slot 12); true
// when there is one or the weapon needs none (flags 0x0A).
func (e *RuleState) AmmoFor(c *Character, it **Item) bool {
	return e.ammoWith(c, c.Slots[SlotWeapon], it)
}

// ammoWith is ammoFor with w as the weapon (extra: the Launcher slot's).
func (e *RuleState) ammoWith(c *Character, w *Item, it **Item) bool {
	*it = nil
	var f byte
	if w != nil {
		f = e.Base(w.Type).Usage
		if f&0x10 != 0 {
			*it = w
		}
		if f&8 != 0 {
			if f&1 != 0 && c.Slots[SlotArrows] != nil && e.isKind(c.Slots[SlotArrows].Type, ItemIDArrow) { // not a missile in the slot
				*it = c.Slots[SlotArrows]
			}
			if f&0x80 != 0 {
				*it = c.Slots[SlotBolts]
			}
		}
	}
	return *it != nil || f == 0xA
}

// Guard = 00BA:2686 (char): end the turn on Guard.
func (e *RuleState) Guard(c *Character) bool {
	r := e.ClearCombatInfo(c)
	c.Combat.Guarding = 1
	e.notice("Guarding")
	return r
}

// Bandage = ovr09:0FFF (doIt): is a party member dying? With doIt the
// first one is bandaged.
func (e *RuleState) Bandage(doIt bool) bool {
	found := false
	for _, c := range e.Party {
		if c.Combat.Extra == 0 && c.Side == SideParty && c.Status == StatusDying {
			found = true
			if doIt {
				c.Status = StatusUnconscious
				c.Combat.Dying = 0
				e.CombatSay(c, "is bandaged", true)
				doIt = false
			}
		}
	}
	return found
}

// QuickOne = ovr09:122B (char): the computer fights for it; a target on
// its own side is dropped.
func (e *RuleState) QuickOne(c *Character) {
	c.Auto = 1
	if t := c.Combat.Target; t != nil && t.Side == c.Side {
		c.Combat.Target = nil
	}
}

// PathNextTo (extra) is the safest walk in reach that brings c, armed for
// melee, next to foe t: fewest free attacks on the way, then fewest foes
// next to where it ends, then fewest steps (nil: t is in reach of its
// weapon already, c shoots, or no square next to t can be reached).
func (e *RuleState) PathNextTo(c, t *Character, reach map[[2]int][]int) []int {
	if e.RangedWeapon(c) || e.RangeTo(c, t) <= 1 {
		return nil
	}
	tx, ty := e.CombatX(t), e.CombatY(t)
	w, h := 1, 1 // t's extent: size 2 is two squares tall, 3 two wide, 4 both
	for part := range 4 {
		if dx, dy, ok := e.Footprint(e.CombatSize(t), part); ok {
			w, h = max(w, dx+1), max(h, dy+1)
		}
	}
	idx := e.CombatIndex(c)
	ox, oy, near, targets := e.Cpos[idx].X, e.Cpos[idx].Y, slices.Clone(e.Near), e.Targets
	defer func() { e.Cpos[idx].X, e.Cpos[idx].Y, e.Near, e.Targets = ox, oy, near, targets }()
	var best []int
	var bestKey [5]int
	found := false
	for sq, p := range reach {
		if sq[0] < tx-1 || sq[0] > tx+w || sq[1] < ty-1 || sq[1] > ty+h {
			continue
		}
		free := 0
		for _, s := range e.WalkSteps(c, p) {
			free += B2i(s.Free)
		}
		e.Cpos[idx].X, e.Cpos[idx].Y = byte(sq[0]), byte(sq[1])
		key := [5]int{free, e.NearEnemies(c, 1), len(p), sq[1], sq[0]} // the square last: map order is random
		e.Cpos[idx].X, e.Cpos[idx].Y = ox, oy
		if !found || slices.Compare(key[:], bestKey[:]) < 0 {
			best, bestKey, found = p, key, true
		}
	}
	if !found {
		return nil
	}
	return append([]int(nil), best...)
}

// loadout (extra): a party member's weapon, shield and Missiles slot items
// and how they were readied, as they were when the computer took over.
type loadout [5]struct {
	it *Item
	r  byte
}

var loadoutSlots = [4]int{SlotWeapon, SlotShield, SlotArrows, SlotBolts}

// saveLoadout (extra) keeps c's loadout the first time the computer picks
// a weapon for it this battle (Quick, charmed ...), for restoreLoadouts.
func (e *RuleState) saveLoadout(c *Character) {
	if !e.partyMember(c) {
		return
	}
	if e.loadouts == nil {
		e.loadouts = map[*Character]*loadout{}
	}
	if e.loadouts[c] != nil {
		return
	}
	var l loadout
	for i, s := range loadoutSlots {
		if it := c.Slots[s]; it != nil {
			l[i].it, l[i].r = it, it.Readied
		}
	}
	e.loadouts[c] = &l
}

// RestoreLoadouts (extra): after the battle everyone the computer fought
// for gets the weapons it had back (those not used up); cursed ones stay.
func (e *RuleState) RestoreLoadouts() {
	for c, l := range e.loadouts {
		keep := func(it *Item) bool {
			for _, x := range l {
				if x.it == it {
					return true
				}
			}
			return false
		}
		for _, s := range loadoutSlots {
			if it := c.Slots[s]; it != nil && it.Cursed == 0 && !keep(it) {
				it.Readied = 0
				if it.Spells[2] > 0x7F {
					e.itemAffect(it, c, 1)
				}
			}
		}
		e.ApplyItems(c)
		for i, x := range l {
			if x.it == nil || x.it.Readied != 0 || c.Slots[loadoutSlots[i]] != nil || !slices.Contains(e.Carried(c), x.it) {
				continue
			}
			e.takeFromPool(c, x.it)
			x.it.Readied = x.r
			if x.it.Spells[2] > 0x7F {
				e.itemAffect(x.it, c, 0)
			}
		}
		e.ApplyItems(c)
	}
	e.loadouts = nil
}
