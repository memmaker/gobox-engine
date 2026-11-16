package rulesxxvc

import (
	"fmt"
	"slices"

	"gobox/rules"
)

// Resolve is the rules outside the board's moves: each case names its .pas.
func (r *Rules) Resolve(s rules.Situation, a rules.Action, rng *rules.RNG) (rules.Outcome, error) {
	rng = r.stream(rng) // the engine passes none: the ruleset's own stream
	if o, ok, err := r.resolveSession(a); ok {
		return o, err
	}
	switch a := a.(type) {
	case rules.OpenTurn:
		// The Pool's menu runs affect group 0xF and disarms a maneuver first; XXVc has neither (its turn-start
		// effects run in BeginTurn): nothing to do.
		return rules.Outcome{}, nil
	case rules.CombatSay: // a line of the fight's text (the bottom line, PrintBottomLine): the battle log's
		text := a.Text
		if c := r.chars[a.Who]; c != nil {
			text = c.Name + " " + text
		}
		return rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: text}}}, nil
	case rules.ClearCombat: // the turn record cleared: no more actions this round
		if s := r.fight[a.Who]; s != nil {
			endActions(s)
		}
		return rules.Outcome{TurnOver: true}, nil
	case rules.FightCleared: // the fight's board and what it laid on it go (BeginFight and EndFight clear the arena)
		r.ground, r.arena, r.fight = rules.Board{}, arena{}, nil
		return rules.Outcome{}, nil
	case rules.ClearTreasure: // the treasure's items go, the pool's credits stay
		clear(r.loot)
		return rules.Outcome{}, nil
	case rules.BankLoot:
		r.bankLoot()
		return rules.Outcome{}, nil
	case rules.EnterPhase:
		return rules.Outcome{}, nil // the engine's phase: no XXVc rule differs by it
	case rules.TakeItems:
		r.takeItems(a)
		return rules.Outcome{}, nil
	case rules.StashItems:
		r.stashItems(a)
		return rules.Outcome{}, nil
	case rules.SkillDamage:
		return r.skillDamage(a, rng), nil
	case rules.StockShop:
		return r.stockShop(a)
	case rules.Buy:
		return r.buy(a.Who, a.Item)
	case rules.Sell:
		return r.sell(a.Who, a.Item, false)
	case rules.Reload:
		return r.reload(a)
	case rules.Confirmed:
		if a, ok := a.Action.(rules.Sell); ok {
			return r.sell(a.Who, a.Item, true)
		}
	case rules.Ready:
		return r.ready(a.Who, a.Item, true)
	case rules.Unready:
		return r.ready(a.Who, a.Item, false)
	}
	switch s := s.(type) {
	case rules.Fight:
		return r.resolveFight(s, a, rng)
	case rules.ShipFight:
		if r.space == nil {
			return rules.Outcome{}, errNoShips
		}
		return r.shipResolve(s, a, rng)
	case rules.Camping:
		if a, ok := a.(rules.Rest); ok {
			return r.rest(s.Party, a.Minutes), nil
		}
	case rules.Looting:
		switch a := a.(type) {
		case rules.TakeItem:
			return r.takeItem(s, a)
		case rules.TakeMoney:
			return r.takeMoney(s, a)
		}
	}
	return rules.Outcome{}, fmt.Errorf("Resolve %T in %T: %w", a, s, ErrNotPorted)
}

// Elapse is gametime.pas AdvanceTime's rule part, TimeOutEffects: effects run
// down for every member (timeOutEffects).
func (r *Rules) Elapse(p rules.Party, minutes int32, _ *rules.RNG) rules.Outcome {
	for _, id := range p {
		if c := r.chars[id]; c != nil {
			timeOutEffects(c, r.fight[id], int(minutes))
		}
	}
	return rules.Outcome{}
}

// Allowed is what the situation offers.
func (r *Rules) Allowed(s rules.Situation) rules.Allowed {
	switch s := s.(type) {
	case rules.Fight:
		return r.allowedFight(s)
	case rules.ShipFight:
		if r.space == nil {
			return rules.Allowed{}
		}
		return r.shipAllowed(s)
	case rules.Camping:
		return rules.Allowed{Offers: []rules.Offer{{Action: rules.Rest{Minutes: MinutesPerDay}, Name: "Rest"}}}
	}
	return rules.Allowed{}
}

// handOut gives each of c's items without one a handle no other item had:
// the item keeps it while the lists change (core.ItemHandle, as rules1e's).
func (r *Rules) handOut(c *Character) {
	for i := range c.Items {
		if c.Items[i].Handle == rules.NoItem {
			r.nextItem++
			c.Items[i].Handle = r.nextItem
		}
	}
}

// heldAt is the index of the item of c that h names, or -1.
func heldAt(c *Character, h rules.ItemHandle) int {
	return slices.IndexFunc(c.Items, func(ci Carried) bool { return ci.Handle == h })
}
