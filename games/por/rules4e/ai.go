package rules4e

import "rules1e"

// The monsters' roles in the computer's play (DMG pp. 54-55):
//   - brutes go for the easiest to hit, soldiers for the best armoured (the
//     front line), skirmishers and lurkers for the most wounded, artillery
//     for the worst armoured, controllers for a spell caster;
//   - brutes, soldiers, minions, elites and solos stand their ground (no
//     morale check);
//   - lurkers act last in the round, striking once the others are engaged;
//   - a soldier marks whom it attacks: until the end of its next turn the mark takes
//     2 more on the number to hit anyone else;
//   - a leader gives the allies within 5 squares +1 to hit.

const (
	markPenalty = 2
	leaderBonus = 1
	leaderRange = 5
)

func monster(c *rules1e.Character) (Monster, bool) {
	if !c.Foe() {
		return Monster{}, false
	}
	return Lookup(c.Name())
}

// Target picks by the attacker's role; nil for a foe the table lacks.
func (r *Rules) Target(_ *rules1e.RuleState, c *rules1e.Character, foes []*rules1e.Character) *rules1e.Character {
	m, ok := monster(c)
	if !ok || len(foes) == 0 {
		return nil
	}
	var score func(f *rules1e.Character) int // the highest is attacked
	switch m.Role {
	case Brute:
		score = func(f *rules1e.Character) int { return -int(f.AC[0]) } // the lowest number to hit
	case Soldier:
		score = func(f *rules1e.Character) int { return int(f.AC[0]) }
	case Artillery:
		score = func(f *rules1e.Character) int { return -int(f.AC[0]) }
	case Controller:
		score = func(f *rules1e.Character) int {
			return 1000*min(f.Level(rules1e.Mage)+f.Level(rules1e.Cleric), 1) - int(f.HP)
		}
	default: // Skirmisher, Lurker
		score = func(f *rules1e.Character) int { return -int(f.HP) }
	}
	best := foes[0]
	for _, f := range foes[1:] {
		if score(f) > score(best) {
			best = f
		}
	}
	return best
}

// Holds: brutes, soldiers, minions, elites and solos never break.
func (r *Rules) Holds(_ *rules1e.RuleState, c *rules1e.Character) bool {
	m, ok := monster(c)
	return ok && (m.Role == Brute || m.Role == Soldier || m.Rank != Standard)
}

// roleRound is a monster's part of RoundStart: a lurker waits for the others.
func (r *Rules) roleRound(c *rules1e.Character) {
	if m, ok := monster(c); ok && m.Role == Lurker && int8(c.Combat.Delay) > 1 {
		c.Combat.Delay = 1 // acts last
	}
}

// roleNeed is the roles' part of HitNeed.
func (r *Rules) roleNeed(e *rules1e.RuleState, a, b *rules1e.Character, need int) int {
	if s := r.marks[a]; s != nil && s != b && s.Okay != 0 && r.has(a, marked) {
		need += markPenalty
	}
	if !a.Foe() {
		return need
	}
	for _, l := range e.Combatants() {
		if m, ok := monster(l); ok && m.Leader && l != a && l.Side == a.Side && e.Distance(l, a) <= leaderRange {
			return need - leaderBonus
		}
	}
	return need
}

// roleAttack is the roles' part of AfterAttack: a soldier marks until
// the end of its next turn, as the Monster Manual's soldiers do.
func (r *Rules) roleAttack(a, b *rules1e.Character) {
	if m, ok := monster(a); ok && m.Role == Soldier && b.Okay != 0 {
		r.marks[b] = a
		r.lay(b, marked, a, true)
	}
}
