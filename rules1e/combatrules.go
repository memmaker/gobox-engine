package rules1e

import (
	"gobox/rules"
	"slices"
)

// Combat rule sets (extra): the decision points where a rule set other than
// the original's may change a fight. ClassicRules is the original's; a
// plug-in (package rules4e) embeds it and overrides what it changes, so a
// new decision point needs no change to the plug-ins. The set can change
// mid-fight (Engine.HandIn): every decision point asks the current one.

// CombatRules are the decision points of a fight.
type CombatRules interface {
	CombatStart(s *RuleState)                       // a fight begins (combatSetup)
	RoundStart(s *RuleState, c *Character)          // a round begins for c (turnSetup: every combatant's, before the first turn)
	TurnStart(s *RuleState, c *Character)           // c's turn begins (takeTurn), skipped or not
	TurnEnd(s *RuleState, c *Character)             // c's turn is over (endTurn)
	Moved(s *RuleState, c *Character)               // c stepped a square in its turn (stepSquare)
	Entered(s *RuleState, c *Character)             // c came into a square, by its own move or not (stepSquare, MoveTo)
	Special(s *RuleState, c *Character) bool        // c's special attack (on-hit poison, paralysis, drain; a gaze, spitting acid) may strike now
	Damage(s *RuleState, c *Character, dmg int) int // the damage c takes in a fight

	// Maneuvers are what c may do besides the original's commands (the
	// bar's Maneuver). One aimed AimWeapon is armed and rides c's next
	// weapon attack (Armed); the others are Used.
	Maneuvers(s *RuleState, c *Character) []Maneuver
	Use(s *RuleState, c *Character, id int, t *Character) (endsTurn bool) // AimSelf (t nil), AimFoe, AimAlly, AimArea; AimWeapon before its attack
	HitNeed(s *RuleState, a, b *Character, need int) int                  // what a's d20 + THAC0 must reach to hit b
	Dealt(s *RuleState, a, b *Character, h Hit) int                       // the damage of a's hit on b, before a backstab's multiple and the damage affects
	Blow(s *RuleState, a *Character, h Hit) Hit                           // the dice and bonus a's weapon hit h deals by this rule set, before a target (the shown damage, DamageText)
	Attacking(s *RuleState, a, b *Character)                              // a's attack on b begins (an interrupt may strike a first)
	Hits(s *RuleState, a, b *Character, over int) bool                    // a's roll hits b by over (0: just); false turns it into a miss (an interrupt), never on a 20
	AfterAttack(s *RuleState, a, b *Character, hit bool)                  // a's attack on b is over
	Provokes(s *RuleState, c, f *Character) bool                          // c leaving foe f's side lets f strike
	Picks(s *RuleState, c *Character) []PowerOwed                         // the powers party member c still has to choose (powers.go)
	Toughness(c *Character) int                                           // the hit points c's powers add to its maximum (Toughness)
	Rested(s *RuleState, c *Character)                                    // c finished a long rest (LongRest)
	Sheet(s *RuleState, c *Character) []Stat                              // the figures the rule set adds to c's sheet (nil: none)

	// The computer's choices.
	Target(s *RuleState, c *Character, foes []*Character) *Character // which of the foes in reach c attacks (nil: the original's random one)
	Holds(s *RuleState, c *Character) bool                           // c stands without a morale check (never flees nor surrenders unless turned)
	OffHand(s *RuleState, c *Character, it *Item) bool               // c may wield weapon it in its off hand, beside the one in hand
	Backstabs(s *RuleState) bool                                     // a thief's attack from behind multiplies its damage (damageRoll)
}

// Hit is a weapon hit's damage roll: Roll on Dice dSides, plus Bonus, of
// which Ability is the strength's part.
type Hit struct {
	Dice, Sides, Roll int
	Bonus, Ability    int
}

// ClassicRules are the original's rules: every decision point leaves the fight as it was.
type ClassicRules struct{}

func (ClassicRules) CombatStart(*RuleState)                                 {}
func (ClassicRules) RoundStart(*RuleState, *Character)                      {}
func (ClassicRules) TurnStart(*RuleState, *Character)                       {}
func (ClassicRules) TurnEnd(*RuleState, *Character)                         {}
func (ClassicRules) Moved(*RuleState, *Character)                           {}
func (ClassicRules) Entered(*RuleState, *Character)                         {}
func (ClassicRules) Attacking(*RuleState, *Character, *Character)           {}
func (ClassicRules) Special(*RuleState, *Character) bool                    { return true }
func (ClassicRules) Hits(*RuleState, *Character, *Character, int) bool      { return true }
func (ClassicRules) Damage(_ *RuleState, _ *Character, d int) int           { return d }
func (ClassicRules) Maneuvers(*RuleState, *Character) []Maneuver            { return nil }
func (ClassicRules) Use(*RuleState, *Character, int, *Character) bool       { return false }
func (ClassicRules) HitNeed(_ *RuleState, _, _ *Character, need int) int    { return need }
func (ClassicRules) Dealt(_ *RuleState, _, _ *Character, h Hit) int         { return h.Roll + h.Bonus }
func (ClassicRules) Blow(_ *RuleState, _ *Character, h Hit) Hit             { return h }
func (ClassicRules) AfterAttack(*RuleState, *Character, *Character, bool)   {}
func (ClassicRules) Provokes(*RuleState, *Character, *Character) bool       { return true }
func (ClassicRules) Picks(*RuleState, *Character) []PowerOwed               { return nil }
func (ClassicRules) Toughness(*Character) int                               { return 0 }
func (ClassicRules) Rested(*RuleState, *Character)                          {}
func (ClassicRules) Sheet(*RuleState, *Character) []Stat                    { return nil }
func (ClassicRules) Target(*RuleState, *Character, []*Character) *Character { return nil }
func (ClassicRules) Holds(*RuleState, *Character) bool                      { return false }
func (ClassicRules) OffHand(*RuleState, *Character, *Item) bool             { return false }
func (ClassicRules) Backstabs(*RuleState) bool                              { return true }

// Settings are the rules' options the frontend sets while the engine runs
// (Engine.HandIn); the rules see them as plain values, taken at a decision
// point (Engine.takeSettings).
type Settings struct {
	Combat     CombatRules // the combat rule set (nil: the original's)
	PathSearch bool        // option (Gameplay "Path search", ruling 3, default off): the computer walks by aiPathDir instead of the original's step
}

// RulesetSettings marks Settings as a ruleset's (core.Configure).
func (Settings) RulesetSettings() {}

// SetRules makes r the combat rules (nil: the original's).
func (e *RuleState) SetRules(r CombatRules) { e.Settings.Combat = r }

// Rules is the current rule set.
func (e *RuleState) Rules() CombatRules {
	if e.Settings.Combat != nil {
		return e.Settings.Combat
	}
	return ClassicRules{}
}

// Armed is the AimWeapon maneuver that rides c's next weapon attack.
type Armed struct {
	C  *Character
	ID int
}

// Armed is the AimWeapon maneuver c's attack carries, 0 for none.
func (e *RuleState) Armed(c *Character) int {
	if e.Fight.Armed.C == c {
		return e.Fight.Armed.ID
	}
	return 0
}

// specialGroups are the affect groups of a monster's special attacks:
// on a hit of slot 1 and 2, and at the start of its turn (gazes, spitting).
var specialGroups = []int{2, 3, 0xE}

// HasSpecial tells whether c has a special attack (an affect of specialGroups).
func (e *RuleState) HasSpecial(c *Character) bool {
	for _, g := range specialGroups {
		for _, t := range affectGroups[g] {
			if e.FindAffect(c, t) != nil {
				return true
			}
		}
	}
	return false
}

// Foe tells whether c fights against the party.
func (c *Character) Foe() bool { return c.Side != SideParty }

// Class is a Levels index.
type Class int

const (
	Cleric  Class = ClsCleric
	Fighter Class = ClsFighter
	Paladin Class = ClsPaladin
	Ranger  Class = ClsRanger
	Mage    Class = ClsMage
	Thief   Class = ClsThief
)

// Level is c's level in class k (0: not of the class).
func (c *Character) Level(k Class) int { return max(int(int8(c.Levels[k])), 0) }

// Human: c is a human.
func (c *Character) Human() bool { return c.Race == RaceHuman }

// Hurt deals dmg plain damage to c, with the message and the fall.
func (e *RuleState) Hurt(c *Character, dmg int) {
	e.Rv.DmgType = 0
	e.Damage(c, dmg, 0, false)
}

// Say shows c's message in the fight's text (and its log).
func (e *RuleState) Say(c *Character, msg string) { e.CombatSay(c, msg, true) }

// Adjacent are the standing foes of c next to it.
func (e *RuleState) Adjacent(c *Character) []*Character {
	var out []*Character
	for i := 1; i <= e.Ncomb; i++ {
		if f := e.Combatant(byte(i)); f != nil && f.Side != c.Side && f.Okay != 0 && e.Cpos[i].Size != 0 && e.RangeTo(c, f) == 1 {
			out = append(out, f)
		}
	}
	return out
}

// Mod is c's 4e modifier for the ability, (score-10)/2 rounded down (PHB p. 17).
func (e *RuleState) Mod(c *Character, a Ability) int { return (int(c.Stats[a]) - 10) >> 1 }

// WeaponAbility is what c's scores add to the THAC0 and damage of the
// weapon it wields (weaponBlow), and whether it shoots.
func (e *RuleState) WeaponAbility(c *Character, with Wield) (hit, dmg int, missile bool) {
	w := with.Weapon
	if w == nil {
		return int(strHit(c)), int(strDamage(c)), false
	}
	flags := e.Base(w.Type).Usage
	if flags&2 != 0 {
		hit, missile = int(rules.DexMissile(byte(c.Stats[StatDex]))), true
	}
	if flags&4 != 0 {
		hit, dmg = hit+int(strHit(c)), int(strDamage(c))
	}
	return
}

// Distance is the squares from a to b (1: next to it); far off (0xFF) when one is not on the map.
func (e *RuleState) Distance(a, b *Character) int {
	if e.CombatIndex(a) == 0 || e.CombatIndex(b) == 0 {
		return 0xFF
	}
	return e.RangeTo(a, b)
}

// Combatants are the fight's standing combatants on the map.
func (e *RuleState) Combatants() []*Character {
	var out []*Character
	for i := 1; i <= e.Ncomb; i++ {
		if c := e.Combatant(byte(i)); c != nil && c.Okay != 0 && e.Cpos[i].Size != 0 {
			out = append(out, c)
		}
	}
	return out
}

// Push moves b one square straight away from a when b fills one square
// and the square beyond is open floor; no free attacks, no moves spent.
func (e *RuleState) Push(a, b *Character) bool {
	return e.RangeTo(a, b) == 1 && e.PushFrom(e.At(a), b)
}

// PushFrom is Push away from square p, wherever b stands.
func (e *RuleState) PushFrom(p rules.BoardPos, b *Character) bool {
	idx := e.CombatIndex(b)
	v := e.At(b).Sub(p)
	if idx == 0 || e.CombatSize(b) != 1 || v == (rules.BoardPos{}) {
		return false
	}
	d := 0
	for d < 8 && (DirDX[d] != rules.Sign(v.X) || DirDY[d] != rules.Sign(v.Y)) {
		d++
	}
	who, ground := e.GroundAt(b, d)
	if ground == 0 || who != 0 || e.T().Ground[ground][0] == 0xFF {
		return false
	}
	ox, oy := e.Cpos[idx].X, e.Cpos[idx].Y
	nx, ny := ox+byte(DirDX[d]), oy+byte(DirDY[d])
	e.Cpos[idx].X, e.Cpos[idx].Y = nx, ny
	if e.Fight.On {
		e.Emit(rules.Walked{Unit: unit(idx), Path: []rules.BoardPos{{X: int(nx), Y: int(ny)}}})
	}
	e.CombatRedrawMap()
	b.Combat.Attackers, b.Combat.Flank = 0, 0
	return true
}

// Aimed is the square an AimArea maneuver was aimed at.
func (e *RuleState) Aimed() rules.BoardPos { return e.Fight.Aimed }

// At is c's square, (-1, -1) off the map.
func (e *RuleState) At(c *Character) rules.BoardPos {
	if e.CombatIndex(c) == 0 {
		return rules.BoardPos{X: -1, Y: -1}
	}
	return rules.BoardPos{X: e.CombatX(c), Y: e.CombatY(c)}
}

// Burst are the standing combatants within r squares of p, walls aside.
func (e *RuleState) Burst(p rules.BoardPos, r int) []*Character {
	var out []*Character
	for _, c := range e.Combatants() {
		if d := e.At(c).Sub(p); max(d.X, -d.X, d.Y, -d.Y) <= r {
			out = append(out, c)
		}
	}
	return out
}

// Shield: c has a shield readied.
func (e *RuleState) Shield(c *Character) bool {
	it := c.Slots[SlotShield]
	return it != nil && e.Base(it.Type).Slot == SlotShield
}

// TwoHanded: the weapon wielded takes both hands.
func (e *RuleState) TwoHanded(w Wield) bool {
	return w.Weapon != nil && e.Base(w.Weapon.Type).Hands > 1
}

// Strike is one weapon attack of a on b outside a's attacks (a power's
// second attack): its number to hit as attackSlots's, its damage dealt.
// w is what a strikes with.
func (e *RuleState) Strike(a, b *Character, w Wield) (hit bool) {
	by, was := e.Rv.By, e.Rv.With
	e.Rv.By, e.Rv.With = a, w
	defer func() { e.Rv.By, e.Rv.With = by, was }()
	need := int(b.AC[0])
	if e.RangedFacing(a, b, w) {
		need = int(b.AC[1]) - 2
	}
	e.RangePenalty(a, b, w, &need)
	need = e.Rules().HitNeed(e, a, b, need)
	mode := 1
	if e.RangedFacing(a, b, w) {
		mode = 2
	}
	if !e.AttackHits(a, b, w, need) {
		e.Sound(string(rules.CueMiss))
		e.attackText(a, b, mode, 0, 0, 0)
		return false
	}
	e.Sound(e.hitEvent(b, "hit"))
	e.DamageRoll(a, b, w, 1)
	e.attackText(a, b, mode, int(e.Rv.Damage), int(e.Rv.Damage), 1)
	return true
}

// Opportunity: the attack being made is a free attack on one leaving its attacker's side.
func (e *RuleState) Opportunity() bool { return e.Fight.free }

// open: (x, y) is floor nobody stands on.
func (e *RuleState) open(x, y int) bool {
	who, g := e.MapAt(x, y)
	return who == 0 && g != 0 && e.GroundInfo(g)[0] != 0xFF
}

// paths are the shortest ways of steps (directions) from c's square to
// each open square within n squares, a diagonal one, round others; nil
// for a combatant larger than a square.
func (e *RuleState) paths(c *Character, n int) map[rules.BoardPos][]int {
	if e.CombatSize(c) != 1 {
		return nil
	}
	from := e.At(c)
	out := map[rules.BoardPos][]int{from: nil}
	for edge := []rules.BoardPos{from}; n > 0 && len(edge) > 0; n-- {
		var next []rules.BoardPos
		for _, p := range edge {
			for d := 0; d < 8; d++ {
				q := p.Add(rules.BoardPos{X: DirDX[d], Y: DirDY[d]})
				if _, seen := out[q]; seen || !e.open(q.X, q.Y) {
					continue
				}
				out[q] = append(slices.Clone(out[p]), d)
				next = append(next, q)
			}
		}
		edge = next
	}
	delete(out, from)
	return out
}

// Squares are the open squares c could get to within n squares (Choose's marks).
func (e *RuleState) Squares(c *Character, n int) []rules.BoardPos {
	var out []rules.BoardPos
	for p := range e.paths(c, n) {
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b rules.BoardPos) int { return a.Y*64 + a.X - b.Y*64 - b.X })
	return out
}

// MoveTo takes c to p (one of Squares) step by step, spending no moves:
// a shift or forced movement. provoke: leaving a foe's side lets it strike
// (a forced move that provokes); false when it did not get there.
func (e *RuleState) MoveTo(c *Character, p rules.BoardPos, provoke bool) bool {
	path, ok := e.paths(c, 0x7F)[p]
	if !ok {
		return false
	}
	e.Fight.forced = true
	defer func() { e.Fight.forced = false }()
	for _, d := range path {
		if provoke {
			e.Fight.forced = false
			e.freeAttacksOnLeave(c, d)
			e.Fight.forced = true
		}
		if c.Okay == 0 || e.IsHelpless(c) || !e.open(e.At(c).X+DirDX[d], e.At(c).Y+DirDY[d]) {
			return false
		}
		e.stepSquare(c, d)
	}
	return e.At(c) == p
}

// Swap switches the squares of a and b, each filling one square.
func (e *RuleState) Swap(a, b *Character) bool {
	i, j := e.CombatIndex(a), e.CombatIndex(b)
	if i == 0 || j == 0 || e.CombatSize(a) != 1 || e.CombatSize(b) != 1 {
		return false
	}
	pa, pb := e.Cpos[i], e.Cpos[j]
	e.Cpos[i].X, e.Cpos[i].Y, e.Cpos[j].X, e.Cpos[j].Y = pb.X, pb.Y, pa.X, pa.Y
	if e.Fight.On {
		e.Emit(rules.Walked{Unit: unit(i), Path: []rules.BoardPos{{X: int(pb.X), Y: int(pb.Y)}}})
		e.Emit(rules.Walked{Unit: unit(j), Path: []rules.BoardPos{{X: int(pa.X), Y: int(pa.Y)}}})
	}
	e.CombatRedrawMap()
	return true
}

// Speed is c's move in squares.
func (e *RuleState) Speed(c *Character) int { return e.MoveRate(c) / 2 }

// Undead: c is undead (turnable).
func (c *Character) IsUndead() bool { return int8(c.Undead) > 0 }

// Controlled: c is a party member the player commands (no hireling, not left to the computer).
func (c *Character) Controlled() bool { return !c.Foe() && c.Control <= 0x7F && c.Auto == 0 }

// WeaponKind is what a rule set tells weapons apart by (the PHB's groups
// and properties, pp. 214-219).
type WeaponKind int

const (
	NoWeapon       WeaponKind = iota
	Dagger                    // a light blade, off-hand
	ShortSword                // a light blade, off-hand
	HandAxe                   // off-hand
	ThrowingHammer            // off-hand
	OneHanded                 // any other one-handed melee weapon
	TwoHandedMelee            // a melee weapon taking both hands
	Bow
	Crossbow
	Sling
	Thrown // darts, javelins, spears thrown
)

// KindOf is it's weapon kind, NoWeapon for no weapon.
func (e *RuleState) KindOf(it *Item) WeaponKind {
	if it == nil {
		return NoWeapon
	}
	switch e.kind(it.Type) {
	case ItemIDDagger:
		return Dagger
	case ItemIDShortSword:
		return ShortSword
	case ItemIDHandAxe:
		return HandAxe
	case ItemIDHammer:
		return ThrowingHammer
	case ItemIDHeavyCrossbow, ItemIDLightCrossbow:
		return Crossbow
	case ItemIDSling:
		return Sling
	case ItemIDDart, ItemIDJavelin:
		return Thrown
	}
	b := e.Base(it.Type)
	switch {
	case e.isBow(it.Type):
		return Bow
	case b.Slot != SlotWeapon:
		return NoWeapon
	case b.Hands > 1:
		return TwoHandedMelee
	}
	return OneHanded
}

// Wielded is the weapon in c's main hand, or with off the one in its off hand (the shield slot); nil for none.
func (e *RuleState) Wielded(c *Character, off bool) *Item {
	if off {
		if it := c.Slots[SlotShield]; it != nil && it.Readied == ReadiedOffHand {
			return it
		}
		return nil
	}
	return c.Slots[SlotWeapon]
}

// Staff: c wields a quarterstaff (an implement, PHB p. 221).
func (e *RuleState) Staff(c *Character) bool {
	it := e.Wielded(c, false)
	return it != nil && e.isKind(it.Type, ItemIDQuarterStaff)
}

// Wand: c has a wand readied (an implement).
func (e *RuleState) Wand(c *Character) bool {
	return slices.ContainsFunc(c.Items, func(it *Item) bool {
		return it.Readied != 0 && e.isKind(it.Type, ItemIDWand, ItemIDWandOfMissiles)
	})
}

// StrikeOff is Strike with the weapon in a's off hand; false without one.
func (e *RuleState) StrikeOff(a, b *Character) bool {
	off := e.Wielded(a, true)
	if off == nil {
		return false
	}
	return e.Strike(a, b, Wield{Weapon: off})
}

// offHand: it, unreadied, would go into c's off hand: a one-handed weapon
// of the weapon hand (not a launcher or thrown-only one: their own slots)
// with one already in hand, the shield slot free, the rule set allowing it.
func (e *RuleState) offHand(c *Character, it *Item) bool {
	w := c.Slots[SlotWeapon]
	return it.Readied == 0 && w != nil && w != it && c.Slots[SlotShield] == nil && e.itemSlot(c, it) == SlotWeapon &&
		e.Base(it.Type).Hands == 1 && e.Base(w.Type).Hands == 1 && e.Rules().OffHand(e, c, it)
}

// Reach is how many squares c's weapon in hand strikes (1: next to it).
func (e *RuleState) Reach(c *Character) int { return max(e.WeaponRange(e.Hand(c)), 1) }

// afflictions are the lasting affects a 4e saving throw may end (Divine Mettle).
var afflictions = []AffectType{AffHeld, affHeldFast, affParalyse2d8, affParalyseNotElf, affParalyse1d9, AffCharmed, affCharmGaze, affSnakeCharm, affFear, AffBlind, AffSlow}

// EndAffliction ends the first of c's afflictions (hold, paralysis, charm,
// fear, blindness, slow); false when it has none.
func (e *RuleState) EndAffliction(c *Character) bool {
	for _, t := range afflictions {
		if e.CureAffect(c, t) {
			return true
		}
	}
	return false
}

// Afflicted: c has an affliction EndAffliction would end.
func (e *RuleState) Afflicted(c *Character) bool {
	return slices.ContainsFunc(afflictions, func(t AffectType) bool { return e.FindAffect(c, t) != nil })
}

// Forced: the square entered (Entered) was a shift or forced movement (MoveTo), no move of its own.
func (e *RuleState) Forced() bool { return e.Fight.forced }
