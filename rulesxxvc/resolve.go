package rulesxxvc

import (
	"fmt"
	"slices"

	"gobox/rules"
)

// Resolve is the rules outside the board's moves: each case names its .pas.
func (r *Rules) Resolve(s rules.Situation, a rules.Action, rng *rules.RNG) (rules.Outcome, error) {
	switch a := a.(type) {
	case rules.EnterPhase:
		return rules.Outcome{}, nil // the engine's phase: no XXVc rule differs by it
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
	case rules.Shopping:
		switch a := a.(type) {
		case rules.Buy:
			g := slices.IndexFunc(s.Goods, func(g rules.ShopItem) bool { return g.Handle == a.Item })
			if g < 0 {
				return rules.Outcome{}, fmt.Errorf("Buy: no goods %d", a.Item)
			}
			return r.buy(a.Who, s.Goods[g], s.Terms)
		case rules.Sell:
			return r.sell(a.Who, a.Item, false)
		case rules.Reload:
			return r.reload(a)
		case rules.Confirmed:
			if a, ok := a.Action.(rules.Sell); ok {
				return r.sell(a.Who, a.Item, true)
			}
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
