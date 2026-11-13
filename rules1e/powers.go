package rules1e

import (
	"errors"
	"gobox/core"
	"slices"
)

// Power picks (extra): the powers a rule set lets party members choose
// (rules4e: 4e's at-will and encounter powers, at the first level and
// later ones). The rule set says what each member still owes (Picks);
// whenever the party is exploring or at the party menu, the first member
// owing one gets the pick screen (View.Powers) until none is owed, so
// turning the rule set on, a level trained, a member joined or a game
// loaded all ask at the next chance. The computer's members (hirelings)
// take the first option. The chosen powers are the member's Powers,
// saved beside its record (NAME.pow).

// PowerOwed is one choice a member owes: one of Options.
type PowerOwed struct {
	What    string // "an at-will power"
	Options []Maneuver
}

// Toughen keeps the rule set's Toughness in c's maximum hit points: it
// comes with the rule set and goes with it.
func (e *RuleState) Toughen(c *Character) {
	if n := e.Rules().Toughness(c) - c.Tough; n != 0 {
		c.MaxHP = byte(int(c.MaxHP) + n)
		c.HP = byte(max(min(int(c.HP)+n, int(c.MaxHP)), min(int(c.HP), 1)))
		c.Tough += n
	}
}

// LongRest is the minutes of rest that bring the daily uses back (Surges,
// Laid): the PHB's extended rest (p. 263).
const LongRest = 6 * 60

// Policy is c's policy for reaction id.
func (c *Character) Policy(id int) Policy { return c.Policies[id] }

// SetPolicy sets c's policy for reaction id.
func (c *Character) SetPolicy(id int, p Policy) {
	if c.Policies == nil {
		c.Policies = map[int]Policy{}
	}
	c.Policies[id] = p % 3
}

// Reactions are the rule set's Reactions c has, if it lists any.
func (e *RuleState) Reactions(c *Character) []core.Reaction {
	r, ok := e.Rules().(interface {
		Reactions(*Character) []core.Reaction
	})
	if !ok {
		return nil
	}
	rs := r.Reactions(c)
	for i := range rs {
		rs[i].Policy = core.Policy(c.Policy(rs[i].ID))
	}
	return rs
}

// setPolicy = SetPolicy for a.Who (core.SetPolicy; Policy counts as core's).
func (r *Rules) setPolicy(a core.SetPolicy) error {
	c := r.chars[a.Who]
	if c == nil {
		return errNoChar
	}
	if !slices.ContainsFunc(r.S.Reactions(c), func(x core.Reaction) bool { return x.ID == a.Reaction }) {
		return errors.New("no such reaction")
	}
	c.SetPolicy(a.Reaction, Policy(a.Policy))
	return nil
}
