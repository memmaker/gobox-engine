package rules4e

import (
	"gobox/rules"
	"gobox/rules1e"
	"slices"
	"strconv"
)

// Players: every class gets the PHB's level-1 shape (pp. 54-55): two
// at-will powers and one encounter power (the ranger: Hunter's Quarry, a
// class feature), and everyone Second Wind (PHB p. 291). Weapon powers
// ride the next attack; the others are aimed or used at once. Spells stay
// as they are, in the place of the daily powers.
//
// A party member's weapon attack is 4e's: the attack's ability modifier
// (Strength in melee, Dexterity shooting, or the power's) takes the place
// of what the original's scores add to the THAC0 and the damage, and a
// power's [W]s roll the weapon's dice (PHB pp. 274-276). An implement
// attack (a spell-like power) takes the THAC0 of the level alone plus its
// ability modifier. Fortitude, Reflex and Will are AC plus defence's
// offset: the Monster Manual's, the DMG's by role, or a character's (PHB p. 274).

// kind is how a member gets a power: a class feature comes with the
// class; at-will and encounter powers are chosen (Picks).
type kind int

const (
	feature kind = iota
	atWill
	encounter
	style // a class feature's option, chosen once at 1st level (group)
)

// ab is an ability a power uses; dflt is the weapon attack's.
type ab int

const (
	dflt ab = iota
	aStr
	aDex
	aWis
	aCha
	aInt
)

func (a ab) of(missile bool) rules1e.Ability {
	switch a {
	case aStr:
		return rules1e.Strength
	case aDex:
		return rules1e.Dexterity
	case aWis:
		return rules1e.Wisdom
	case aCha:
		return rules1e.Charisma
	case aInt:
		return rules1e.Intelligence
	}
	if missile {
		return rules1e.Dexterity
	}
	return rules1e.Strength
}

// none is a hit with no ability modifier (1[W] alone).
var none = []ab{}

// power is one maneuver and its rules.
type power struct {
	rules1e.Maneuver
	classes []rules1e.Class // nil: every class
	kind    kind
	group   string // a style's feature
	once    bool   // once a fight (encounter powers, Second Wind)

	// a weapon power's attack
	atk    ab                                                                // its ability
	vs     def                                                               // the defence it attacks
	hit    func(r *Rules, e *rules1e.RuleState, a *rules1e.Character) int    // its bonus to hit
	w      int                                                               // its [W]s; 0: 1
	abl    []ab                                                              // the damage's ability modifiers; nil: the attack's
	more   func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) int // its damage beyond the [W]s and abilities
	after  func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character)     // hit or miss
	onHit  func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character)
	onMiss func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character)

	before  func(r *Rules, e *rules1e.RuleState, c *rules1e.Character)      // a weapon power's move before its attack
	trigger bool                                                            // used when its trigger comes, not from the bar (Fox's Cunning)
	req     func(e *rules1e.RuleState, c *rules1e.Character) bool           // what c must wield; nil: anything
	can     func(r *Rules, e *rules1e.RuleState, c *rules1e.Character) bool // c may use it now (a surge left, uses left); nil: yes
	perRnd  bool                                                            // once a round
	uses    int                                                             // times a fight, beside once (Healing Word: 2)
	channel bool                                                            // Channel Divinity: one of them a fight
	perTurn bool                                                            // once a turn
	blast   bool                                                            // AimArea: a close blast of Area*2+1 squares, its side next to c

	use func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) (endsTurn bool)
}

var (
	fighter, paladin, ranger = rules1e.Fighter, rules1e.Paladin, rules1e.Ranger
	cleric, mage, thief      = rules1e.Cleric, rules1e.Mage, rules1e.Thief
)

// basic is the basic attack's (PHB p. 287).
var basic = &power{}

var powers = []power{
	{Maneuver: m(1, "Second Wind", "Spend a healing surge: heal a quarter of your hit points; +2 defence until your next turn. Once a fight.", rules1e.AimSelf, 0), once: true,
		can: func(_ *Rules, e *rules1e.RuleState, c *rules1e.Character) bool { return surgesLeft(e, c) > 0 },
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			surge(e, c, 0)
			r.lay(c, guard2, c, false)
			e.Say(c, "catches a second wind")
			return true
		}},

	// fighter (PHB pp. 77-78)
	{Maneuver: m(2, "Cleave", "On a melee hit, another foe next to you takes your Strength modifier in damage.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: atWill,
		onHit: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			if e.Distance(a, b) != 1 {
				return
			}
			for _, f := range e.Adjacent(a) {
				if f != b {
					if d := e.Mod(a, rules1e.Strength); d > 0 {
						e.Hurt(f, d)
					}
					return
				}
			}
		}},
	{Maneuver: m(3, "Sure Strike", "+2 to hit; no ability modifier on the damage.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: atWill,
		hit: plus(2), abl: none},
	{Maneuver: m(4, "Steel Serpent Strike", "2[W], and the foe is slowed (2 squares) until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: encounter, once: true,
		w: 2,
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			r.lay(b, slowed, a, true)
			e.Say(b, "is slowed")
		}},

	// paladin (PHB p. 92)
	{Maneuver: m(5, "Valiant Strike", "+1 to hit for each foe next to you.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: atWill,
		hit: func(_ *Rules, e *rules1e.RuleState, a *rules1e.Character) int { return len(e.Adjacent(a)) }},
	{Maneuver: m(6, "Bolstering Strike", "Charisma attack and damage; on a hit, temporary hit points equal to your Wisdom modifier.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: atWill,
		atk: aCha, abl: []ab{aCha},
		onHit: func(r *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			r.shield(a, e.Mod(a, rules1e.Wisdom))
		}},
	{Maneuver: m(7, "Radiant Smite", "2[W] + Strength and Wisdom modifiers. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: encounter, once: true,
		w: 2, abl: []ab{aStr, aWis}},

	// ranger (PHB pp. 104-105)
	{Maneuver: m(8, "Careful Attack", "+2 to hit; no ability modifier on the damage.", rules1e.AimWeapon, 0), classes: []rules1e.Class{ranger}, kind: atWill,
		hit: plus(2), abl: none},
	{Maneuver: m(9, "Hit and Run", "Leaving the foe's side this turn gives it no free attack.", rules1e.AimWeapon, 0), classes: []rules1e.Class{ranger}, kind: atWill,
		after: func(r *Rules, _ *rules1e.RuleState, a, b *rules1e.Character) { r.hitRun[a] = b }},
	{Maneuver: m(10, "Hunter's Quarry", "The nearest foe is your quarry: once a round, a hit on it deals 1d6 more. Does not end the turn.", rules1e.AimSelf, 0), classes: []rules1e.Class{ranger},
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			var q *rules1e.Character
			for _, f := range e.Combatants() {
				if f.Side != c.Side && (q == nil || e.Distance(c, f) < e.Distance(c, q)) {
					q = f
				}
			}
			if q != nil {
				r.quarry[c] = q
				e.Say(q, "is the quarry")
			}
			return false
		}},

	// cleric (PHB pp. 62-63)
	{Maneuver: m(11, "Priest's Shield", "On a hit, you and an ally next to you get +1 defence until the end of your next turn.", rules1e.AimWeapon, 0), classes: []rules1e.Class{cleric}, kind: atWill,
		onHit: func(r *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			r.lay(a, guard1, a, true)
			for _, f := range e.Combatants() {
				if f != a && f.Side == a.Side && e.Distance(a, f) == 1 {
					r.lay(f, guard1, a, true)
					return
				}
			}
		}},
	{Maneuver: m(12, "Sacred Flame", "Wisdom against Reflex, 5 squares: 1d6 + Wisdom modifier, and the most wounded ally gains temporary hit points.", rules1e.AimFoe, 5), classes: []rules1e.Class{cleric}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "calls down sacred flame")
			if !r.attack(e, c, t, rules1e.Wisdom, ref) {
				return true
			}
			e.Hurt(t, max(e.Dice(1, 6)+e.Mod(c, rules1e.Wisdom), 0))
			var ally *rules1e.Character
			for _, f := range e.Combatants() {
				if f != c && f.Side == c.Side && (ally == nil || f.MaxHP-f.HP > ally.MaxHP-ally.HP) {
					ally = f
				}
			}
			if ally != nil { // the other choice, a saving throw, has nothing to end: no condition here is "save ends"
				r.shield(ally, e.Mod(c, rules1e.Charisma)+c.Level(cleric)/2)
			}
			return true
		}},
	{Maneuver: m(13, "Wrathful Thunder", "On a hit, the foe is dazed (one action a turn, no free attacks) until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{cleric}, kind: encounter, once: true,
		onHit: daze},

	// thief (PHB pp. 117-118)
	{Maneuver: m(14, "Piercing Strike", "With a light blade: Dexterity against Reflex, Dexterity on the damage.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: atWill,
		req: lightBlade,
		atk: aDex, vs: ref, abl: []ab{aDex}},
	{Maneuver: m(15, "Sly Flourish", "With a light blade, crossbow or sling: Dexterity attack; Dexterity and Charisma on the damage.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: atWill,
		req: rogueWeapon,
		atk: aDex, abl: []ab{aDex, aCha}},
	{Maneuver: m(16, "Dazing Strike", "With a light blade: Dexterity attack and damage; on a hit, the foe is dazed (one action a turn, no free attacks) until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: encounter, once: true,
		req: lightBlade,
		atk: aDex, abl: []ab{aDex}, onHit: daze},

	// mage (PHB pp. 158-159)
	{Maneuver: m(17, "Magic Missile", "Intelligence against Reflex, 20 squares: 2d4 + Intelligence modifier.", rules1e.AimFoe, 20), classes: []rules1e.Class{mage}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "looses a magic missile")
			r.bolt(e, c, t, ref, 2, 4)
			return true
		}},
	{Maneuver: m(18, "Ray of Frost", "Intelligence against Fortitude, 10 squares: 1d6 + Intelligence modifier, and the foe is slowed (2 squares) until the end of your next turn.", rules1e.AimFoe, 10), classes: []rules1e.Class{mage}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "casts a ray of frost")
			if r.bolt(e, c, t, fort, 1, 6) {
				r.lay(t, slowed, c, true)
			}
			return true
		}},
	{Maneuver: m(19, "Force Orb", "Intelligence against Reflex, 20 squares: 2d8 + Intelligence modifier; on a hit, each of its allies next to it is attacked for 1d10 + Intelligence modifier. Once a fight.", rules1e.AimFoe, 20), classes: []rules1e.Class{mage}, kind: encounter, once: true,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "hurls a force orb")
			if r.bolt(e, c, t, ref, 2, 8) {
				for _, f := range e.Combatants() {
					if f != t && f.Side == t.Side && e.Distance(t, f) == 1 {
						r.bolt(e, c, f, ref, 1, 10)
					}
				}
			}
			return true
		}},
	// fighter
	{Maneuver: m(20, "Reaping Strike", "On a miss, half your Strength modifier in damage (all of it with a two-handed weapon).", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: atWill,
		onMiss: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			d := e.Mod(a, rules1e.Strength)
			if !e.TwoHanded(a) {
				d /= 2
			}
			if d > 0 {
				e.Hurt(b, d)
			}
		}},
	{Maneuver: m(21, "Tide of Iron", "With a shield: on a hit, you push the foe 1 square and can shift into its square.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: atWill,
		req: (*rules1e.RuleState).Shield,
		onHit: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			was := e.At(b)
			if e.Push(a, b) {
				if a.Controlled() { // into its square
					e.MoveTo(a, was, false)
				}
			}
		}},
	{Maneuver: m(22, "Spinning Sweep", "On a hit, the foe is knocked prone (standing up takes its move). Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: encounter, once: true,
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			r.lay(b, prone, nil, false)
			e.Say(b, "is knocked prone")
		}},
	{Maneuver: m(23, "Passing Attack", "On a hit, you can shift 1 square, then a second attack at +2 on another foe next to you. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: encounter, once: true,
		hit: func(r *Rules, _ *rules1e.RuleState, a *rules1e.Character) int {
			if r.second[a] {
				return 2
			}
			return 0
		},
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			shift(e, a, 1, b)
			var others []*rules1e.Character
			for _, f := range e.Adjacent(a) {
				if f != b {
					others = append(others, f)
				}
			}
			if f := pick(others); f != nil {
				r.second[a] = true
				e.Strike(a, f)
				delete(r.second, a)
			}
		}},

	// paladin
	{Maneuver: m(24, "Divine Challenge", "Mark a foe within 5 squares: -2 to its attacks on others, and the first each turn burns it for 3 + your Charisma modifier. Lasts while you attack it or end your turn next to it; when it lapses, none on your next turn. Once a turn; does not end the turn.", rules1e.AimFoe, 5), classes: []rules1e.Class{paladin}, perTurn: true,
		can: func(r *Rules, _ *rules1e.RuleState, c *rules1e.Character) bool { return !r.lapsed[c] },
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			for f, p := range r.chal {
				if p == c {
					delete(r.chal, f)
				}
			}
			r.chal[t], r.marks[t] = c, c
			r.lay(t, marked, c, true)
			e.Say(t, "is challenged")
			return false
		}},
	{Maneuver: m(25, "Enfeebling Strike", "Charisma attack and damage; on a hit, a foe you marked takes -2 to its attacks until the end of your next turn.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: atWill,
		atk: aCha, abl: []ab{aCha},
		onHit: func(r *Rules, _ *rules1e.RuleState, a, b *rules1e.Character) {
			if r.marks[b] == a && r.has(b, marked) {
				r.lay(b, feeble, a, true)
			}
		}},
	{Maneuver: m(26, "Holy Strike", "Your Wisdom modifier in more damage on a foe you marked.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: atWill,
		more: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) int {
			if r.marks[b] == a && r.has(b, marked) {
				return e.Mod(a, rules1e.Wisdom)
			}
			return 0
		}},
	{Maneuver: m(27, "Piercing Smite", "Against Reflex, 2[W]; on a hit, the foe and as many foes next to you as your Wisdom modifier are marked until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: encounter, once: true,
		vs: ref, w: 2,
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			n := e.Mod(a, rules1e.Wisdom)
			for _, f := range append([]*rules1e.Character{b}, e.Adjacent(a)...) {
				if f != b && n <= 0 {
					break
				}
				if f != b {
					n--
				}
				r.marks[f] = a
				r.lay(f, marked, a, true)
			}
		}},

	// ranger
	{Maneuver: m(28, "Twin Strike", "With a bow or crossbow, or two melee weapons (main hand, then off hand): two attacks, on one foe or two, 1[W] each, no ability modifier.", rules1e.AimWeapon, 0), classes: []rules1e.Class{ranger}, kind: atWill,
		req: twoAttacks, abl: none,
		after: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) { second(e, a, b) }},
	{Maneuver: m(29, "Dire Wolverine Strike", "With two melee weapons: an attack on each foe next to you, 1[W] + Strength modifier. Once a fight.", rules1e.AimSelf, 0), classes: []rules1e.Class{ranger}, kind: encounter, once: true,
		req: twoBlades,
		use: func(_ *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "whirls like a dire wolverine")
			for _, f := range e.Adjacent(c) {
				e.Strike(c, f)
			}
			return true
		}},
	{Maneuver: m(39, "Two-Fanged Strike", "With a bow or crossbow, or two melee weapons: two attacks on the foe, 1[W] + ability modifier each; both hit: your Wisdom modifier more. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{ranger}, kind: encounter, once: true,
		req: twoAttacks,
		onHit: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			if b.Okay == 0 {
				return
			}
			hit := false
			if missile(e, a) {
				hit = e.Strike(a, b)
			} else {
				hit = e.StrikeOff(a, b)
			}
			if hit && b.Okay != 0 {
				if d := e.Mod(a, rules1e.Wisdom); d > 0 {
					e.Hurt(b, d)
				}
			}
		},
		onMiss: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			if missile(e, a) {
				e.Strike(a, b)
			} else {
				e.StrikeOff(a, b)
			}
		}},

	// cleric
	{Maneuver: m(30, "Lance of Faith", "Wisdom against Reflex, 5 squares: 1d8 + Wisdom modifier, and an ally's next attack on the foe gets +2.", rules1e.AimFoe, 5), classes: []rules1e.Class{cleric}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "casts a lance of faith")
			if r.attack(e, c, t, rules1e.Wisdom, ref) {
				e.Hurt(t, max(e.Dice(1, 8)+e.Mod(c, rules1e.Wisdom), 0))
				r.lance[t] = c
			}
			return true
		}},
	{Maneuver: m(31, "Divine Glow", "A blast 3 squares wide next to you: Wisdom against Reflex on each foe, 1d8 + Wisdom modifier; allies in it get +2 to attack until the end of your next turn. Once a fight.", rules1e.AimArea, 2), classes: []rules1e.Class{cleric}, kind: encounter, once: true,
		blast: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "calls a divine glow")
			for _, t := range e.Burst(e.Aimed(), 1) {
				if t.Side == c.Side {
					r.lay(t, glow, c, true)
				} else if r.attack(e, c, t, rules1e.Wisdom, ref) {
					e.Hurt(t, max(e.Dice(1, 8)+e.Mod(c, rules1e.Wisdom), 0))
				}
			}
			return true
		}},

	// thief
	{Maneuver: m(32, "Torturous Strike", "With a light blade: Dexterity attack, 2[W] + Dexterity modifier (Brutal Scoundrel: and Strength). Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: encounter, once: true,
		atk: aDex, w: 2, abl: []ab{aDex}, req: lightBlade,
		more: func(_ *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) int {
			if picked(a, brutalScoundrel) {
				return e.Mod(a, rules1e.Strength)
			}
			return 0
		}},

	// mage
	{Maneuver: m(33, "Scorching Burst", "Intelligence against Reflex on each creature within 1 square of a square 10 away: 1d6 + Intelligence modifier.", rules1e.AimArea, 10), classes: []rules1e.Class{mage}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "casts a scorching burst")
			for _, t := range e.Burst(e.Aimed(), 1) {
				r.bolt(e, c, t, ref, 1, 6)
			}
			return true
		}},
	{Maneuver: m(34, "Thunderwave", "A blast 3 squares wide next to you: Intelligence against Fortitude on each creature, 1d6 + Intelligence modifier, and a push of your Wisdom modifier in squares.", rules1e.AimArea, 2), classes: []rules1e.Class{mage}, kind: atWill,
		blast: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "casts thunderwave")
			for _, t := range e.Burst(e.Aimed(), 1) {
				if r.bolt(e, c, t, fort, 1, 6) {
					for range e.Mod(c, rules1e.Wisdom) {
						if !pushAway(e, c, t) {
							break
						}
					}
				}
			}
			return true
		}},
	{Maneuver: m(35, "Burning Hands", "A blast 5 squares wide next to you: Intelligence against Reflex on each creature, 2d6 + Intelligence modifier. Once a fight.", rules1e.AimArea, 3), classes: []rules1e.Class{mage}, kind: encounter, once: true,
		blast: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "casts burning hands")
			for _, t := range e.Burst(e.Aimed(), 2) {
				r.bolt(e, c, t, ref, 2, 6)
			}
			return true
		}},
	{Maneuver: m(36, "Chill Strike", "Intelligence against Fortitude, 10 squares: 2d8 + Intelligence modifier, and the foe is dazed until the end of your next turn. Once a fight.", rules1e.AimFoe, 10), classes: []rules1e.Class{mage}, kind: encounter, once: true,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "casts chill strike")
			if r.bolt(e, c, t, fort, 2, 8) {
				daze(r, e, c, t)
			}
			return true
		}},
	{Maneuver: m(37, "Ray of Enfeeblement", "Intelligence against Fortitude, 10 squares: 1d10 + Intelligence modifier, and the foe is weakened (half damage) until the end of your next turn. Once a fight.", rules1e.AimFoe, 10), classes: []rules1e.Class{mage}, kind: encounter, once: true,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "casts ray of enfeeblement")
			if r.bolt(e, c, t, fort, 1, 10) {
				r.lay(t, weakened, c, true)
				e.Say(t, "is weakened")
			}
			return true
		}},
	// fighter
	{Maneuver: m(38, "Covering Attack", "2[W]; on a hit, an ally next to the foe can shift 2 squares. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{fighter}, kind: encounter, once: true,
		w: 2,
		onHit: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			var near []*rules1e.Character
			for _, f := range e.Adjacent(b) {
				if f != a {
					near = append(near, f)
				}
			}
			if f := pick(near); f != nil && a.Controlled() {
				shift(e, f, 2, nil)
			}
		}},

	// paladin
	{Maneuver: m(40, "Fearsome Smite", "Charisma attack, 2[W] + Charisma modifier; on a hit, the foe takes your Wisdom modifier less to its attacks until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: encounter, once: true,
		atk: aCha, w: 2, abl: []ab{aCha},
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			r.layN(b, fear, a, max(e.Mod(a, rules1e.Wisdom), 0))
		}},
	{Maneuver: m(41, "Shielding Smite", "Charisma attack, 2[W] + Charisma modifier; hit or miss, an ally within 5 squares gets your Wisdom modifier to AC until the end of your next turn. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{paladin}, kind: encounter, once: true,
		atk: aCha, w: 2, abl: []ab{aCha},
		after: func(r *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			allies := within(e, a, 5, true)
			slices.SortFunc(allies, func(x, y *rules1e.Character) int { return int(x.HP) - int(y.HP) }) // the computer's pick: the most hurt
			if f := pick(allies); f != nil {
				r.layN(f, ward, a, max(e.Mod(a, rules1e.Wisdom), 0))
			}
		}},

	// ranger
	{Maneuver: m(42, "Nimble Strike", "With a bow or crossbow: shift 1 square before or after the attack.", rules1e.AimWeapon, 0), classes: []rules1e.Class{ranger}, kind: atWill,
		req:    missile,
		before: func(r *Rules, e *rules1e.RuleState, c *rules1e.Character) { r.shifted[c] = shift(e, c, 1, nil) },
		after: func(r *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			if !r.shifted[a] {
				shift(e, a, 1, nil)
			}
		}},
	{Maneuver: m(43, "Evasive Strike", "2[W]; shift 1 + your Wisdom modifier in squares before or after the attack. Once a fight.", rules1e.AimWeapon, 0, 4), classes: []rules1e.Class{ranger}, kind: encounter, once: true,
		w: 2,
		before: func(r *Rules, e *rules1e.RuleState, c *rules1e.Character) {
			// the player's square first; the computer's missile shifts out of
			// reach before, a melee attack after
			r.shifted[c] = aimedMove(e, c, evasive(e, c), false) || !c.Controlled() && missile(e, c) && shift(e, c, evasive(e, c), nil)
		},
		after: func(r *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			if !r.shifted[a] {
				shift(e, a, evasive(e, a), nil)
			}
		}},
	{Maneuver: m(foxsCunning, "Fox's Cunning", "When a foe attacks you in melee: shift 1 square, then a basic attack on it with your Wisdom modifier to hit. As its policy on the sheet says; once a fight.", rules1e.AimSelf, 0), classes: []rules1e.Class{ranger}, kind: encounter, once: true,
		trigger: true},

	// cleric
	{Maneuver: m(45, "Righteous Brand", "On a hit, an ally within 5 squares gets your Strength modifier to melee attacks on the foe until the end of your next turn.", rules1e.AimWeapon, 0), classes: []rules1e.Class{cleric}, kind: atWill,
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			if f := pick(within(e, a, 5, true)); f != nil {
				r.brand[b] = brand{f, max(e.Mod(a, rules1e.Strength), 0)}
				r.lay(b, branded, a, true)
			}
		}},
	{Maneuver: m(46, "Cause Fear", "Wisdom against Will, 10 squares: the foe runs its speed + your Charisma modifier away from you, leaving its foes' sides as it goes. Once a fight.", rules1e.AimFoe, 10), classes: []rules1e.Class{cleric}, kind: encounter, once: true,
		use: func(r *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "causes fear")
			if r.attack(e, c, t, rules1e.Wisdom, will) {
				var far rules.BoardPos
				d := -1
				for _, p := range e.Squares(t, e.Speed(t)+max(e.Mod(c, rules1e.Charisma), 0)) {
					if v := p.Sub(e.At(c)); max(v.X, -v.X, v.Y, -v.Y) > d {
						far, d = p, max(v.X, -v.X, v.Y, -v.Y)
					}
				}
				if d >= 0 {
					e.Say(t, "flees in fear")
					e.MoveTo(t, far, true)
				}
			}
			return true
		}},

	// thief
	{Maneuver: m(48, "Deft Strike", "With a light blade, crossbow or sling: Dexterity attack and damage; you can move 2 squares first.", rules1e.AimWeapon, 0, 2), classes: []rules1e.Class{thief}, kind: atWill,
		req: rogueWeapon,
		atk: aDex, abl: []ab{aDex},
		before: func(_ *Rules, e *rules1e.RuleState, c *rules1e.Character) { aimedMove(e, c, 2, true) }, // the computer's: none
	},
	{Maneuver: m(49, "Riposte Strike", "With a light blade: Dexterity attack and damage; if the foe attacks you before your next turn, you strike it first (Strength, 1[W] + Strength modifier).", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: atWill,
		req: lightBlade,
		atk: aDex, abl: []ab{aDex},
		onHit: func(r *Rules, _ *rules1e.RuleState, a, b *rules1e.Character) { r.riposte[b] = a }},
	{Maneuver: m(50, "King's Castle", "With a light blade, crossbow or sling: Dexterity against Reflex, 2[W] + Dexterity modifier; hit or miss, you switch places with an ally next to you. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: encounter, once: true,
		req: rogueWeapon,
		atk: aDex, vs: ref, w: 2, abl: []ab{aDex},
		after: func(_ *Rules, e *rules1e.RuleState, a, _ *rules1e.Character) {
			if !a.Controlled() {
				return
			}
			if f := pick(within(e, a, 1, true)); f != nil {
				e.Swap(a, f)
			}
		}},
	{Maneuver: m(51, "Positioning Strike", "With a light blade: Dexterity against Will, 1[W] + Dexterity modifier; on a hit, you slide the foe 1 square (Artful Dodger: your Charisma modifier). Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{thief}, kind: encounter, once: true,
		req: lightBlade,
		atk: aDex, vs: will, abl: []ab{aDex},
		onHit: func(_ *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			n := 1
			if picked(a, artfulDodger) {
				n = max(e.Mod(a, rules1e.Charisma), 1)
			}
			if a.Controlled() { // the foe away from the party
				shiftFoe(e, b, n)
			}
		}},
	// healing (PHB pp. 92, 62)
	{Maneuver: m(47, "Healing Strike", "2[W] + Strength modifier; on a hit, the foe is marked until the end of your next turn, and you or an ally within 5 squares can spend a healing surge. Once a fight.", rules1e.AimWeapon, 0), classes: []rules1e.Class{cleric}, kind: encounter, once: true,
		w: 2,
		onHit: func(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
			r.marks[b] = a
			r.lay(b, marked, a, true)
			cs := append([]*rules1e.Character{a}, within(e, a, 5, true)...)
			slices.SortStableFunc(cs, func(x, y *rules1e.Character) int { return int(y.MaxHP-y.HP) - int(x.MaxHP-x.HP) }) // the computer's: the most hurt
			if f := pick(cs); f != nil {
				surge(e, f, 0)
			}
		}},
	{Maneuver: m(60, "Healing Word", "You or an ally within 5 squares spends a healing surge and heals 1d6 + your Wisdom modifier more. Twice a fight, once a round; does not end the turn.", rules1e.AimAlly, 5), classes: []rules1e.Class{cleric}, uses: 2, perRnd: true,
		use: func(_ *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "speaks a healing word")
			surge(e, t, e.Dice(1, 6)+max(e.Mod(c, rules1e.Wisdom), 0)) // Healer's Lore
			return false
		}},
	{Maneuver: m(61, "Lay on Hands", "Spend one of your healing surges: an ally next to you heals as if it spent one. Your Wisdom modifier times a day, once a round; does not end the turn.", rules1e.AimAlly, 1), classes: []rules1e.Class{paladin}, perRnd: true,
		can: func(_ *Rules, e *rules1e.RuleState, c *rules1e.Character) bool {
			return surgesLeft(e, c) > 0 && c.Laid < max(e.Mod(c, rules1e.Wisdom), 1)
		},
		use: func(_ *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			c.Surges++
			c.Laid++
			e.Say(c, "lays on hands")
			e.Heal(t, max(int(t.MaxHP)/4, 1), false)
			return false
		}},

	// Channel Divinity: one a fight (PHB pp. 62, 91)
	{Maneuver: m(62, "Divine Strength", "Channel Divinity: your Strength modifier more on your next attack this turn. Does not end the turn.", rules1e.AimSelf, 0), classes: []rules1e.Class{paladin}, channel: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			r.strength[c] = true
			e.Say(c, "channels divine strength")
			return false
		}},
	{Maneuver: m(63, "Divine Mettle", "Channel Divinity: one within 10 squares makes a saving throw with your Charisma modifier against a hold, paralysis, charm, fear, blindness or slow. Does not end the turn.", rules1e.AimAlly, 10), classes: []rules1e.Class{paladin}, channel: true,
		use: func(_ *Rules, e *rules1e.RuleState, c, t *rules1e.Character) bool {
			e.Say(c, "channels divine mettle")
			if !e.Afflicted(t) || e.Dice(1, 20)+e.Mod(c, rules1e.Charisma) < 10 { // a saving throw: 10 or more (PHB p. 279)
				e.Say(t, "is not freed")
				return false
			}
			e.EndAffliction(t)
			return false
		}},
	{Maneuver: m(64, "Divine Fortune", "Channel Divinity: +1 to your next attack before the end of your next turn. Does not end the turn.", rules1e.AimSelf, 0), classes: []rules1e.Class{cleric}, channel: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			r.lay(c, fortune, c, true)
			e.Say(c, "channels divine fortune")
			return false
		}},
	{Maneuver: m(65, "Turn Undead", "Channel Divinity: Wisdom against Will on each undead within 2 squares: 1d10 + Wisdom modifier, a push of 3 + your Charisma modifier squares, and it cannot move until the end of your next turn; a miss: half damage.", rules1e.AimSelf, 0), classes: []rules1e.Class{cleric}, channel: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "channels the power to turn undead")
			for _, t := range within(e, c, 2, false) {
				if !t.IsUndead() {
					continue
				}
				d := max(e.Dice(1, 10)+e.Mod(c, rules1e.Wisdom), 0)
				if !r.attack(e, c, t, rules1e.Wisdom, will) {
					e.Hurt(t, d/2)
					continue
				}
				e.Hurt(t, d)
				for range 3 + max(e.Mod(c, rules1e.Charisma), 0) {
					if !pushAway(e, c, t) {
						break
					}
				}
				r.lay(t, immobile, c, true)
			}
			return true
		}},
	// mage zones
	{Maneuver: m(52, "Cloud of Daggers", "Intelligence against Reflex on each creature in a square 10 away: 1d6 + Intelligence modifier; until the end of your next turn, who enters it or starts its turn there takes your Wisdom modifier (at least 1).", rules1e.AimArea, 10), classes: []rules1e.Class{mage}, kind: atWill,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "conjures a cloud of daggers")
			for _, t := range e.Burst(e.Aimed(), 0) {
				r.bolt(e, c, t, ref, 1, 6)
			}
			r.zones = append(r.zones, zone{at: e.Aimed(), by: c, dmg: max(e.Mod(c, rules1e.Wisdom), 1)})
			return true
		}},
	{Maneuver: m(53, "Icy Terrain", "Intelligence against Reflex on each creature within 1 square of a square 10 away: 1d6 + Intelligence modifier and knocked prone; the ice is difficult terrain until the end of your next turn. Once a fight.", rules1e.AimArea, 10), classes: []rules1e.Class{mage}, kind: encounter, once: true,
		use: func(r *Rules, e *rules1e.RuleState, c, _ *rules1e.Character) bool {
			e.Say(c, "casts icy terrain")
			for _, t := range e.Burst(e.Aimed(), 1) {
				if r.bolt(e, c, t, ref, 1, 6) {
					r.lay(t, prone, nil, false)
				}
			}
			r.zones = append(r.zones, zone{at: e.Aimed(), r: 1, by: c})
			return true
		}},
}

// surgesPerDay is c's healing surges a day: its class's (the highest
// level's) and its Constitution modifier (PHB chapter 4).
func surgesPerDay(e *rules1e.RuleState, c *rules1e.Character) int {
	best, n := 0, 6
	for k, s := range map[rules1e.Class]int{fighter: 9, paladin: 10, ranger: 6, cleric: 7, thief: 6, mage: 6} {
		if l := c.Level(k); l > best || l == best && l > 0 && s > n {
			best, n = l, s
		}
	}
	return max(n+e.Mod(c, rules1e.Constitution), 1)
}

// surgesLeft is how many healing surges party member c has left today; none for a foe.
func surgesLeft(e *rules1e.RuleState, c *rules1e.Character) int {
	if c.Foe() {
		return 0
	}
	return max(surgesPerDay(e, c)-c.Surges, 0)
}

// surge spends one of c's healing surges: a quarter of its hit points and
// more (PHB p. 293); false when it has none left.
func surge(e *rules1e.RuleState, c *rules1e.Character, more int) bool {
	if surgesLeft(e, c) == 0 {
		e.Say(c, "has no healing surge left")
		return false
	}
	c.Surges++
	e.Heal(c, max(int(c.MaxHP)/4, 1)+more, false)
	return true
}

// foxsCunning is Fox's Cunning's id; it fires on its trigger (react.go).
const foxsCunning = 44

// the class features' options (style), chosen at 1st level
const (
	oneHanded = 100 + iota
	twoHanded
	artfulDodger
	brutalScoundrel
	archer
	twoBlade
	staffMastery
	wandMastery
)

// styles are the class features' options (PHB pp. 76, 104, 117, 158); they
// work by themselves, none is on the bar.
var styles = []power{
	{Maneuver: m(oneHanded, "One-Handed Talent", "+1 to hit with a one-handed weapon.", rules1e.AimSelf, 0), classes: []rules1e.Class{fighter}, kind: style, group: "Fighter Weapon Talent"},
	{Maneuver: m(twoHanded, "Two-Handed Talent", "+1 to hit with a two-handed weapon.", rules1e.AimSelf, 0), classes: []rules1e.Class{fighter}, kind: style, group: "Fighter Weapon Talent"},
	{Maneuver: m(artfulDodger, "Artful Dodger", "Your Charisma modifier to AC against opportunity attacks; Positioning Strike slides the foe that many squares.", rules1e.AimSelf, 0), classes: []rules1e.Class{thief}, kind: style, group: "Rogue Tactics"},
	{Maneuver: m(brutalScoundrel, "Brutal Scoundrel", "Your Strength modifier more on Sneak Attack and Torturous Strike.", rules1e.AimSelf, 0), classes: []rules1e.Class{thief}, kind: style, group: "Rogue Tactics"},
	{Maneuver: m(archer, "Archer Style", "Defensive Mobility: +2 to AC against opportunity attacks.", rules1e.AimSelf, 0), classes: []rules1e.Class{ranger}, kind: style, group: "Fighting Style"},
	{Maneuver: m(twoBlade, "Two-Blade Style", "Any one-handed weapon in the off hand, and Toughness: 5 more hit points.", rules1e.AimSelf, 0), classes: []rules1e.Class{ranger}, kind: style, group: "Fighting Style"},
	{Maneuver: m(staffMastery, "Staff of Defense", "With a quarterstaff: +1 to your defences, and once a fight, when hit, your Constitution modifier more against that attack.", rules1e.AimSelf, 0), classes: []rules1e.Class{mage}, kind: style, group: "Arcane Implement Mastery"},
	{Maneuver: m(wandMastery, "Wand of Accuracy", "With a wand readied: once a fight, your Dexterity modifier to hit with a spell.", rules1e.AimSelf, 0), classes: []rules1e.Class{mage}, kind: style, group: "Arcane Implement Mastery"},
}

func init() { powers = append(powers, styles...) }

// picked: c chose option id of a class feature (and has the class).
func picked(c *rules1e.Character, id int) bool {
	i := slices.IndexFunc(styles, func(p power) bool { return p.ID == id })
	return slices.Contains(c.Powers, id) && slices.ContainsFunc(styles[i].classes, func(k rules1e.Class) bool { return c.Level(k) > 0 })
}

// Toughness: 5 hit points come with the Two-Blade style (PHB p. 201).
func (r *Rules) Toughness(c *rules1e.Character) int {
	if picked(c, twoBlade) {
		return 5
	}
	return 0
}

// Rested: a long rest heals fully and gives the daily uses back (PHB p. 263).
func (r *Rules) Rested(e *rules1e.RuleState, c *rules1e.Character) {
	c.Surges, c.Laid = 0, 0
	e.Heal(c, int(c.MaxHP), false)
}

// Sheet: the defences (PHB p. 274; the AC is the sheet's own), the
// healing surges left and what one heals, Lay on Hands left.
func (r *Rules) Sheet(e *rules1e.RuleState, c *rules1e.Character) []rules1e.Stat {
	if c.Foe() {
		return nil
	}
	n := strconv.Itoa
	st := func(name, v string, help ...string) rules1e.Stat {
		return rules1e.Stat{Name: name, Value: v, Help: help}
	}
	ac := int(c.AC[0]) - 40 // the 4e AC: defence's base
	out := []rules1e.Stat{
		st("Fortitude", n(ac+defence(e, c, fort)), "Defence against poison,", "disease and brute force.", "", "10 + armour + the better", "of STR and CON modifier", "+ class bonus."),
		st("Reflex", n(ac+defence(e, c, ref)), "Defence against breath,", "blasts and traps.", "", "10 + armour + the better", "of DEX and INT modifier", "+ class bonus."),
		st("Will", n(ac+defence(e, c, will)), "Defence against charms,", "fear and illusions.", "", "10 + armour + the better", "of WIS and CHA modifier", "+ class bonus."),
		st("Healing Surges", n(surgesLeft(e, c))+"/"+n(surgesPerDay(e, c)), "Surges left today / per day.", "Second Wind and healing", "spend one. A long rest", "restores them all.", "", "Class base + CON modifier."),
		st("Surge Value", n(max(int(c.MaxHP)/4, 1)), "Hit points one healing", "surge restores.", "", "A quarter of maximum HP."),
	}
	if c.Level(paladin) > 0 {
		k := max(e.Mod(c, rules1e.Wisdom), 1)
		out = append(out, st("Lay on Hands", n(max(k-c.Laid, 0))+"/"+n(k), "Uses left today / per day:", "spend a healing surge to", "heal an ally by touch.", "", "WIS modifier, at least 1."))
	}
	return out
}

// once spends c's once-a-fight use of feature id; false when spent.
func (r *Rules) once(c *rules1e.Character, id int) bool {
	if r.used[c][id] {
		return false
	}
	if r.used[c] == nil {
		r.used[c] = map[int]bool{}
	}
	r.used[c][id] = true
	return true
}

// lightBlade: c wields a light blade (a dagger, a short sword).
func lightBlade(e *rules1e.RuleState, c *rules1e.Character) bool {
	k := e.KindOf(e.Wielded(c, false))
	return k == rules1e.Dagger || k == rules1e.ShortSword
}

// rogueWeapon: c wields a light blade, a crossbow or a sling.
func rogueWeapon(e *rules1e.RuleState, c *rules1e.Character) bool {
	k := e.KindOf(e.Wielded(c, false))
	return lightBlade(e, c) || k == rules1e.Crossbow || k == rules1e.Sling
}

// twoBlades: c wields a melee weapon in each hand.
func twoBlades(e *rules1e.RuleState, c *rules1e.Character) bool {
	return e.Wielded(c, true) != nil && !missile(e, c)
}

// twoAttacks: c wields a ranged weapon or two melee weapons.
func twoAttacks(e *rules1e.RuleState, c *rules1e.Character) bool {
	return missile(e, c) || twoBlades(e, c)
}

// second is a two-attack power's second attack, on b or another foe in
// reach: the off hand's in melee.
func second(e *rules1e.RuleState, a, b *rules1e.Character) {
	var foes []*rules1e.Character
	if b.Okay != 0 {
		foes = append(foes, b)
	}
	for _, f := range within(e, a, e.Reach(a), false) {
		if f != b {
			foes = append(foes, f)
		}
	}
	f := pick(foes)
	switch {
	case f == nil:
	case missile(e, a):
		e.Strike(a, f)
	default:
		e.StrikeOff(a, f)
	}
}

// OffHand: anyone may wield a dagger, a short sword, a hand axe or a
// throwing hammer in the off hand (PHB pp. 215-219: off-hand weapons); a
// ranger of the Two-Blade style any one-handed weapon (PHB p. 104).
func (r *Rules) OffHand(e *rules1e.RuleState, c *rules1e.Character, it *rules1e.Item) bool {
	switch e.KindOf(it) {
	case rules1e.Dagger, rules1e.ShortSword, rules1e.HandAxe, rules1e.ThrowingHammer:
		return true
	case rules1e.OneHanded:
		return slices.Contains(c.Powers, twoBlade) && c.Level(ranger) > 0
	}
	return false
}

// missile: c wields a bow or crossbow.
func missile(e *rules1e.RuleState, c *rules1e.Character) bool {
	_, _, m := e.WeaponAbility(c)
	return m
}

// talent is a party member's weapon's own bonus to hit b: Fighter Weapon
// Talent, Rogue Weapon Talent (a dagger), Prime Shot.
func (r *Rules) talent(e *rules1e.RuleState, a, b *rules1e.Character, missile bool) int {
	n := 0
	if picked(a, oneHanded) && !e.TwoHanded(a) && e.Wielded(a, false) != nil || picked(a, twoHanded) && e.TwoHanded(a) {
		n++
	}
	if a.Level(thief) > 0 && e.KindOf(e.Wielded(a, false)) == rules1e.Dagger {
		n++
	}
	if a.Level(ranger) > 0 && missile && !slices.ContainsFunc(within(e, b, e.Distance(a, b)-1, true), func(f *rules1e.Character) bool { return f.Side == a.Side }) {
		n++ // Prime Shot: no ally nearer the target
	}
	return n
}

// advantage: a has combat advantage against b (PHB p. 279): b dazed, prone
// (in melee), flanked by a and an ally on its far side, or not yet acted
// in the fight against a rogue (First Strike).
func (r *Rules) advantage(e *rules1e.RuleState, a, b *rules1e.Character) bool {
	melee := e.Distance(a, b) == 1
	if r.has(b, dazed) || r.has(b, prone) && melee || a.Level(thief) > 0 && !a.Foe() && !r.acted[b] {
		return true
	}
	if !melee {
		return false
	}
	far := e.At(b).Mul(2).Sub(e.At(a))
	for _, f := range e.Combatants() {
		if f != a && f.Side == a.Side && e.At(f) == far {
			return true
		}
	}
	return false
}

// sneak: a's hit on b gets a rogue's Sneak Attack (PHB p. 117): once a
// round, with combat advantage, a light blade, a crossbow or a sling.
func (r *Rules) sneak(e *rules1e.RuleState, a, b *rules1e.Character) bool {
	if a.Foe() || a.Level(thief) == 0 || r.sneaked[a] || !r.advantage(e, a, b) {
		return false
	}
	switch e.KindOf(e.Wielded(a, false)) {
	case rules1e.Dagger, rules1e.ShortSword, rules1e.Crossbow, rules1e.Sling:
		return true
	}
	return false
}

// Backstabs: no backstab multiple; Sneak Attack takes its place.
func (r *Rules) Backstabs(*rules1e.RuleState) bool { return false }

// plus is a constant bonus to hit.
func plus(n int) func(*Rules, *rules1e.RuleState, *rules1e.Character) int {
	return func(*Rules, *rules1e.RuleState, *rules1e.Character) int { return n }
}

// pushAway pushes t a square straight away from c, wherever it stands.
func pushAway(e *rules1e.RuleState, c, t *rules1e.Character) bool {
	return e.PushFrom(e.At(c), t)
}

// daze dazes b until the end of a's next turn.
func daze(r *Rules, e *rules1e.RuleState, a, b *rules1e.Character) {
	r.lay(b, dazed, a, true)
	e.Say(b, "is dazed")
}

// attack is c's implement attack with the ability against t's defence d:
// the level's THAC0 (HitDice) and the ability modifier.
func (r *Rules) attack(e *rules1e.RuleState, c, t *rules1e.Character, a rules1e.Ability, d def) bool {
	if pol := c.Policy(wandMastery); picked(c, wandMastery) && e.Wand(c) && !r.used[c][wandMastery] && pol != rules1e.Never {
		k := max(e.Mod(c, rules1e.Dexterity), 0)
		if pol == rules1e.OnHit { // spent in Hits, only when k made the hit
			r.wandTry[c] = k
			defer delete(r.wandTry, c)
		} else {
			r.once(c, wandMastery)
			e.Say(c, "aims the wand")
		}
		r.bonus[c] = k
		defer delete(r.bonus, c)
	}
	need := r.need(e, c, t, int(t.AC[0])+defence(e, t, d)) + int(c.THAC0) - int(c.Thac0Base) - e.Mod(c, a)
	if !e.AttackHits(c, t, need) {
		e.Say(c, "misses")
		return false
	}
	return true
}

// bolt is a mage's attack on t against d: on a hit, n dSides + Intelligence modifier.
func (r *Rules) bolt(e *rules1e.RuleState, c, t *rules1e.Character, d def, n, sides int) bool {
	if !r.attack(e, c, t, rules1e.Intelligence, d) {
		return false
	}
	e.Hurt(t, max(e.Dice(n, sides)+e.Mod(c, rules1e.Intelligence), 0))
	return true
}

// shield gives c n temporary hit points; they do not add up, the most stays (PHB p. 293).
func (r *Rules) shield(c *rules1e.Character, n int) {
	r.temp[c] = max(r.temp[c], n)
}

func m(id int, name, tip string, aim rules1e.Aim, rng int, move ...int) rules1e.Maneuver {
	mv := rules1e.Maneuver{ID: id, Name: name, Tip: tip, Aim: aim, Range: rng}
	if len(move) > 0 {
		mv.Move = move[0] // the most it moves first (aimedMove checks the member's own reach)
	}
	return mv
}

// evasive is Evasive Strike's shift: 1 + Wisdom modifier squares.
func evasive(e *rules1e.RuleState, c *rules1e.Character) int {
	return 1 + max(e.Mod(c, rules1e.Wisdom), 0)
}

// aimedMove takes a member the player runs to the square aimed at before
// the power (Maneuver.Move), when it is another within n squares; a
// shift unless provoke. False: it stays.
func aimedMove(e *rules1e.RuleState, c *rules1e.Character, n int, provoke bool) bool {
	p := e.Aimed()
	if !c.Controlled() || p == e.At(c) || !slices.Contains(e.Squares(c, n), p) {
		return false
	}
	return e.MoveTo(c, p, provoke)
}

// powerOf is the power with the maneuver id, nil when none.
func powerOf(id int) *power {
	i := slices.IndexFunc(powers, func(p power) bool { return p.ID == id })
	if i < 0 {
		return nil
	}
	return &powers[i]
}

// Maneuvers are a party member's class features and chosen powers of
// its classes (Picks), those once a fight while unused.
func (r *Rules) Maneuvers(e *rules1e.RuleState, c *rules1e.Character) []rules1e.Maneuver {
	if c.Foe() {
		return nil
	}
	var out []rules1e.Maneuver
	for _, p := range powers {
		if r.ofClass(c, &p) && (p.kind == feature || slices.Contains(c.Powers, p.ID)) && !r.used[c][p.ID] && !r.turned[c][p.ID] && !p.trigger && p.kind != style && (p.req == nil || p.req(e, c)) && (p.can == nil || p.can(r, e, c)) &&
			!(p.perRnd && r.rounded[c][p.ID]) && !(p.uses > 0 && r.count[c][p.ID] >= p.uses) && !(p.channel && r.channeled[c]) {
			mv := p.Maneuver
			if p.ID == 43 { // Evasive Strike: the aim reaches its real shift
				mv.Move = evasive(e, c)
			}
			out = append(out, mv)
		}
	}
	return out
}

// ofClass: p is of one of c's classes.
func (r *Rules) ofClass(c *rules1e.Character, p *power) bool {
	return p.classes == nil || slices.ContainsFunc(p.classes, func(k rules1e.Class) bool { return c.Level(k) > 0 })
}

var classNames = map[rules1e.Class]string{fighter: "fighter", paladin: "paladin", ranger: "ranger", cleric: "cleric", mage: "mage", thief: "thief"}

// owed is how many powers of the kind a member of the level knows (PHB
// pp. 27-28): two at-will powers, a human three (PHB p. 46); an encounter
// power at 1st, 3rd and 7th level. A human's extra at-will power is of its
// highest class.
func owed(k kind, level int, human bool) int {
	if k == atWill {
		if human {
			return 3
		}
		return 2
	}
	n := 1
	if level >= 3 {
		n++
	}
	if level >= 7 {
		n++
	}
	return n
}

// Picks are the at-will and encounter powers a party member owes, class by class.
func (r *Rules) Picks(_ *rules1e.RuleState, c *rules1e.Character) []rules1e.PowerOwed {
	if c.Foe() {
		return nil
	}
	best := rules1e.Class(-1)
	for _, k := range []rules1e.Class{fighter, paladin, ranger, cleric, mage, thief} {
		if c.Level(k) > 0 && (best < 0 || c.Level(k) > c.Level(best)) {
			best = k
		}
	}
	var out []rules1e.PowerOwed
	for _, k := range []rules1e.Class{fighter, paladin, ranger, cleric, mage, thief} {
		if c.Level(k) == 0 {
			continue
		}
		groups := map[string][]rules1e.Maneuver{}
		var order []string
		for _, p := range styles {
			if !slices.Contains(p.classes, k) {
				continue
			}
			if _, ok := groups[p.group]; !ok {
				order = append(order, p.group)
			}
			groups[p.group] = append(groups[p.group], p.Maneuver)
		}
		for _, g := range order {
			if !slices.ContainsFunc(groups[g], func(m rules1e.Maneuver) bool { return slices.Contains(c.Powers, m.ID) }) {
				out = append(out, rules1e.PowerOwed{What: "the " + classNames[k] + "'s " + g, Options: groups[g]})
			}
		}
		for _, kd := range []kind{atWill, encounter} {
			var have int
			var free []rules1e.Maneuver
			for _, p := range powers {
				if p.kind != kd || !slices.Contains(p.classes, k) {
					continue
				}
				if slices.Contains(c.Powers, p.ID) {
					have++
				} else {
					free = append(free, p.Maneuver)
				}
			}
			if have < owed(kd, c.Level(k), c.Human() && k == best) && len(free) > 0 {
				what := map[kind]string{atWill: "an at-will", encounter: "an encounter"}[kd]
				out = append(out, rules1e.PowerOwed{What: what + " " + classNames[k] + " power", Options: free})
			}
		}
	}
	return out
}

// spend marks c's power p used when it is once a fight or a turn.
func (r *Rules) spend(c *rules1e.Character, p *power) {
	if p.perRnd {
		if r.rounded[c] == nil {
			r.rounded[c] = map[int]bool{}
		}
		r.rounded[c][p.ID] = true
	}
	if p.uses > 0 {
		if r.count[c] == nil {
			r.count[c] = map[int]int{}
		}
		r.count[c][p.ID]++
	}
	if p.channel {
		r.channeled[c] = true
	}
	if p.perTurn {
		if r.turned[c] == nil {
			r.turned[c] = map[int]bool{}
		}
		r.turned[c][p.ID] = true
	}
	if !p.once {
		return
	}
	if r.used[c] == nil {
		r.used[c] = map[int]bool{}
	}
	r.used[c][p.ID] = true
}

func (r *Rules) Use(e *rules1e.RuleState, c *rules1e.Character, id int, t *rules1e.Character) bool {
	p := powerOf(id)
	if p != nil && p.Aim == rules1e.AimWeapon {
		if p.before != nil {
			p.before(r, e, c)
		}
		return false
	}
	if p == nil || p.use == nil {
		return false
	}
	if d := e.Aimed().Sub(e.At(c)); p.blast && max(d.X, -d.X, d.Y, -d.Y) != p.Area+1 {
		e.Say(c, "aims the blast next to itself")
		return false // a close blast's side is next to its caster (PHB p. 272)
	}
	r.spend(c, p)
	return p.use(r, e, c, t)
}

// weapon is a's weapon power, the basic attack's when none is armed; nil for a foe.
func (r *Rules) weapon(e *rules1e.RuleState, a *rules1e.Character) *power {
	if a.Foe() {
		return nil
	}
	if p := powerOf(e.Armed(a)); p != nil {
		return p
	}
	return basic
}

func (r *Rules) HitNeed(e *rules1e.RuleState, a, b *rules1e.Character, need int) int {
	if p := r.weapon(e, a); p != nil {
		abHit, _, missile := e.WeaponAbility(a)
		need += abHit - e.Mod(a, p.atk.of(missile)) + defence(e, b, p.vs)
		if p.hit != nil {
			need -= p.hit(r, e, a)
		}
		need -= r.talent(e, a, b, missile)
	}
	return r.need(e, a, b, need)
}

// need is HitNeed less the weapon's part: the defender's guard and
// condition, and the roles.
func (r *Rules) need(e *rules1e.RuleState, a, b *rules1e.Character, need int) int {
	switch {
	case r.has(b, guard2):
		need += 2
	case r.has(b, guard1):
		need++
	}
	if r.advantage(e, a, b) {
		need -= 2 // combat advantage (PHB p. 279)
	}
	if r.has(b, prone) && e.Distance(a, b) > 1 {
		need += 2 // PHB p. 277
	}
	if e.Opportunity() && !b.Foe() {
		if picked(b, artfulDodger) {
			need += max(e.Mod(b, rules1e.Charisma), 0)
		}
		if picked(b, archer) {
			need += 2 // Defensive Mobility (PHB p. 196)
		}
	}
	if picked(b, staffMastery) && e.Staff(b) {
		need++
	}
	if r.has(a, feeble) {
		need += 2
	}
	if r.has(a, glow) {
		need -= 2
	}
	if r.has(a, fortune) {
		need--
	}
	if g := r.lance[b]; g != nil && g != a && g.Side == a.Side {
		need -= 2
	}
	need += r.conds[a][fear].n + r.conds[b][ward].n - r.bonus[a]
	if br, ok := r.brand[b]; ok && br.ally == a && r.has(b, branded) && e.Distance(a, b) == 1 {
		need -= br.n
	}
	if e.Opportunity() && a.Level(fighter) > 0 && !a.Foe() {
		need -= max(e.Mod(a, rules1e.Wisdom), 0) // Combat Superiority
	}
	return r.roleNeed(e, a, b, need)
}

func (r *Rules) Dealt(e *rules1e.RuleState, a, b *rules1e.Character, h rules1e.Hit) int {
	dmg := h.Roll + h.Bonus
	if p := r.weapon(e, a); p != nil {
		dmg = weaponDamage(e, a, p, h)
		if p.more != nil {
			dmg += p.more(r, e, a, b)
		}
	}
	if r.quarry[a] == b && !r.quarried[a] {
		r.quarried[a] = true
		dmg += e.Dice(1, 6)
	}
	if r.strength[a] {
		delete(r.strength, a)
		dmg += max(e.Mod(a, rules1e.Strength), 0)
	}
	if r.sneak(e, a, b) {
		r.sneaked[a] = true
		dmg += e.Dice(2, 6)
		if picked(a, brutalScoundrel) {
			dmg += max(e.Mod(a, rules1e.Strength), 0)
		}
	}
	if r.has(a, weakened) {
		dmg /= 2 // PHB p. 277
	}
	return max(dmg, 0)
}

// weaponDamage is p's damage with h: its [W]s, the weapon's own bonus and
// magic, and its abilities' modifiers in place of the original's strength.
func weaponDamage(e *rules1e.RuleState, a *rules1e.Character, p *power, h rules1e.Hit) int {
	dmg := h.Roll + h.Bonus - h.Ability
	for range max(p.w, 1) - 1 {
		dmg += e.Dice(h.Dice, h.Sides)
	}
	abl := p.abl
	if abl == nil {
		abl = []ab{p.atk}
	}
	_, _, missile := e.WeaponAbility(a)
	for _, x := range abl {
		dmg += e.Mod(a, x.of(missile))
	}
	return dmg
}

func (r *Rules) AfterAttack(e *rules1e.RuleState, a, b *rules1e.Character, hit bool) {
	r.roleAttack(a, b)
	r.react(e, a, b, hit)
	delete(r.conds[a], fortune) // its next attack
	delete(r.strength, a)
	if g := r.lance[b]; g != nil && g != a && g.Side == a.Side {
		delete(r.lance, b) // its next attack
	}
	if r.chal[b] == a {
		r.engaged[a] = true
	}
	if p := r.chal[a]; p != nil && p != b && r.marks[a] == p && r.has(a, marked) && !r.stung[a] {
		r.stung[a] = true
		e.Say(a, "is seared by the challenge")
		e.Hurt(a, 3+e.Mod(p, rules1e.Charisma))
	}
	if r.has(a, dazed) && a.Combat != nil {
		a.Combat.Moves = 0 // its one action
	}
	p := powerOf(e.Armed(a))
	if p == nil {
		return
	}
	r.spend(a, p)
	if p.after != nil {
		p.after(r, e, a, b)
	}
	if hit && p.onHit != nil {
		p.onHit(r, e, a, b)
	}
	if !hit && p.onMiss != nil {
		p.onMiss(r, e, a, b)
	}
}

// Provokes: no free attack from a dazed foe (PHB p. 277), nor from the foe of a Hit and Run.
func (r *Rules) Provokes(_ *rules1e.RuleState, c, f *rules1e.Character) bool {
	return r.hitRun[c] != f && !r.has(f, dazed)
}
