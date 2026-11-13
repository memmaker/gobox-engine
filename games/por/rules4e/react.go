package rules4e

import (
	"gobox/rules"
	"gobox/rules1e"
	"slices"
)

// Immediate actions (PHB p. 268) and the moves powers make. An interrupt
// or reaction that costs nothing more than the immediate action fires by
// itself (Combat Challenge, a riposte); one that spends a power follows the
// member's standing policy (rules.Policy, set on the sheet: Reactions):
// nothing asks mid-turn. A power's "you can shift" moves a member the
// player runs to the square farthest from its foes (shift); the aim at one
// of several is the computer's pick (pick); a member the computer runs
// stays where it is.

// shift lets c shift up to n squares (no free attacks): to where its
// nearest foe is farthest, within reach of keep when it is given; false
// for a member the computer runs, or when no square is better.
func shift(e *rules1e.RuleState, c *rules1e.Character, n int, keep *rules1e.Character) bool {
	return c.Controlled() && slide(e, c, n, keep)
}

// shiftFoe slides foe b up to n squares away from the party.
func shiftFoe(e *rules1e.RuleState, b *rules1e.Character, n int) { slide(e, b, n, nil) }

// slide is shift's move, for anyone.
func slide(e *rules1e.RuleState, c *rules1e.Character, n int, keep *rules1e.Character) bool {
	room := func(p rules.BoardPos) int { // the distance to the nearest foe
		d := 1 << 30
		for _, f := range within(e, c, 1<<30, false) {
			d = min(d, cheb(p, e.At(f)))
		}
		return d
	}
	best := e.At(c)
	d := room(best)
	for _, p := range e.Squares(c, n) {
		if keep != nil && cheb(p, e.At(keep)) > e.Reach(c) {
			continue
		}
		if v := room(p); v > d {
			best, d = p, v
		}
	}
	return best != e.At(c) && e.MoveTo(c, best, false)
}

// cheb is the squares between p and q (diagonals count one).
func cheb(p, q rules.BoardPos) int {
	v := p.Sub(q)
	return max(v.X, -v.X, v.Y, -v.Y)
}

// pick is the first of cs (the callers order them the computer's way);
// nil when none.
func pick(cs []*rules1e.Character) *rules1e.Character {
	if len(cs) == 0 {
		return nil
	}
	return cs[0]
}

// within are c's side's combatants (allies: not c) or its foes within n squares of c.
func within(e *rules1e.RuleState, c *rules1e.Character, n int, allies bool) []*rules1e.Character {
	var out []*rules1e.Character
	for _, f := range e.Combatants() {
		if f != c && (f.Side == c.Side) == allies && e.Distance(c, f) <= n {
			out = append(out, f)
		}
	}
	return out
}

// immediate spends c's immediate action of the round; false when spent
// (one a round, never in its own turn: PHB p. 268).
func (r *Rules) immediate(c *rules1e.Character) bool {
	if r.reacted[c] || c.Okay == 0 {
		return false
	}
	r.reacted[c] = true
	return true
}

// Attacking: the interrupts on a's attack on b. A fighter's marked foe next
// to it that attacks another is struck first (Combat Challenge); a foe that
// attacks the rogue whose Riposte Strike hit it takes the riposte.
func (r *Rules) Attacking(e *rules1e.RuleState, a, b *rules1e.Character) {
	if f := r.marks[a]; f != nil && f != b && f.Level(fighter) > 0 && f.Side != a.Side && r.has(a, marked) && e.Distance(f, a) == 1 && r.immediate(f) {
		e.Say(f, "answers the challenge")
		e.Strike(f, a)
	}
	if t := r.riposte[a]; t == b && a.Okay != 0 && e.Distance(a, b) == 1 && r.immediate(b) {
		delete(r.riposte, a)
		e.Say(b, "ripostes")
		e.Strike(b, a)
	}
}

// Hits: a wizard hit while wielding a quarterstaff may ward with Staff of
// Defense, its Constitution modifier more on the defence (PHB p. 158);
// only when that turns the hit into a miss (Always and OnHit alike), never
// under Never. A Wand of Accuracy held for a hit (attack) is spent only
// when its bonus made this one.
func (r *Rules) Hits(e *rules1e.RuleState, a, b *rules1e.Character, over int) bool {
	wand := func() bool { // spent only once the hit stands
		if k, ok := r.wandTry[a]; ok && over < k && r.once(a, wandMastery) {
			e.Say(a, "aims the wand")
		}
		return true
	}
	if b.Side == a.Side || !picked(b, staffMastery) || !e.Staff(b) || b.Okay == 0 || r.reacted[b] || e.Mod(b, rules1e.Constitution) <= over || b.Policy(staffMastery) == rules1e.Never {
		return wand()
	}
	if !r.once(b, staffMastery) {
		return wand()
	}
	r.immediate(b)
	e.Say(b, "wards with the staff")
	return false
}

// Reactions are c's reactions that spend a power, which its policy answers.
func (r *Rules) Reactions(c *rules1e.Character) []rules.Reaction {
	var out []rules.Reaction
	if fox := powerOf(foxsCunning); !c.Foe() && slices.Contains(c.Powers, fox.ID) && r.ofClass(c, fox) {
		out = append(out, rules.Reaction{ID: fox.ID, Name: "Fox"})
	}
	if picked(c, staffMastery) {
		out = append(out, rules.Reaction{ID: staffMastery, Name: "Staff"})
	}
	if picked(c, wandMastery) {
		out = append(out, rules.Reaction{ID: wandMastery, Name: "Wand"})
	}
	return out
}

// react is AfterAttack's part: the reactions to a's attack on b, and the
// fighter's marks and Combat Superiority.
func (r *Rules) react(e *rules1e.RuleState, a, b *rules1e.Character, hit bool) {
	if a.Level(fighter) > 0 && !a.Foe() && b.Okay != 0 && b.Side != a.Side { // Combat Challenge: every attack marks, not over an ally's mark
		if m := r.marks[b]; m == nil || m == a || !r.has(b, marked) || m.Side != a.Side {
			r.marks[b] = a
			r.lay(b, marked, a, true)
		}
	}
	if e.Opportunity() && hit && a.Level(fighter) > 0 && b.Combat != nil {
		b.Combat.Moves = 0 // Combat Superiority: it stops moving
	}
	fox := powerOf(foxsCunning)
	if !b.Foe() && slices.Contains(b.Powers, fox.ID) && r.ofClass(b, fox) && !r.used[b][fox.ID] && !r.reacted[b] && b.Okay != 0 && a.Okay != 0 && e.Distance(a, b) == 1 {
		if pol := b.Policy(fox.ID); pol == rules1e.Never || pol == rules1e.OnHit && !hit {
			return
		}
		r.immediate(b)
		r.spend(b, fox)
		e.Say(b, "uses fox's cunning")
		shift(e, b, 1, a)
		r.bonus[b] = max(e.Mod(b, rules1e.Wisdom), 0)
		e.Strike(b, a)
		delete(r.bonus, b)
	}
}
