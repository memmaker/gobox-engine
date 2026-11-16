package rulesxxvc

import (
	"errors"
	"slices"

	"gobox/rules"
)

// The gear menu's item verbs (viewchar.pas UseItemMenu): Trade gives an item to a member the player picks, Halve
// splits a stack, Join takes in the others of its kind (the stack limit is maxStack, the carry limit
// maxCarriedItems; shop.go joinItem).

var errNoItem = errors.New("rulesxxvc: the member carries no such item")

// The gear menu's words (viewchar.pas PrintBottomLine, in UseItemMenu's actions).
const (
	gearOverloaded  = "Overloaded"        // ItemTradeMenu: the other carries MAX_CARRIED_ITEMS already
	gearUnready     = "Must be unreadied" // CheckItemAvail
	gearCannotHalve = "Can't halve that"  // HalveItem: not joinable, or a stack of 1
)

// handlingOffers are the gear menu's verbs for c's item (UseItemMenu): Trade for a player character or a disabled
// one, not in a fight or on a ship; Halve while the list has room for the half; Join always.
func (r *Rules) handlingOffers(who rules.CharID, h rules.ItemHandle) rules.Allowed {
	c := r.chars[who]
	if c == nil || heldAt(c, h) < 0 {
		return rules.Allowed{}
	}
	var out []rules.Offer
	if (!c.NonPlayer || !c.Enabled) && r.phase != rules.PhaseFight && r.space == nil {
		out = append(out, rules.Offer{Action: rules.GiveItem{From: who, To: rules.NoChar, Item: h}, Name: "Trade"})
	}
	if len(c.Items) < maxCarriedItems {
		out = append(out, rules.Offer{Action: rules.Halve{Who: who, Item: h}, Name: "Halve"})
	}
	return rules.Allowed{Offers: append(out, rules.Offer{Action: rules.JoinStack{Who: who, Item: h}, Name: "Join"})}
}

// refuseGear is a gear verb the original answers with a line on the bottom row.
func refuseGear(text string) (rules.Outcome, error) {
	return rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: text}}}, rules.ErrNotNow
}

// gearItem is the member and item index an action names.
func (r *Rules) gearItem(who rules.CharID, h rules.ItemHandle) (*Character, int, error) {
	c := r.chars[who]
	if c == nil {
		return nil, 0, errNoMember
	}
	i := heldAt(c, h)
	if i < 0 {
		return nil, 0, errNoItem
	}
	return c, i, nil
}

// halveItem is HalveItem: half the quantity (rounded down) becomes a new item beside it, unreadied; not for an item
// that does not join or holds less than two.
func (r *Rules) halveItem(who rules.CharID, h rules.ItemHandle) (rules.Outcome, error) {
	c, i, err := r.gearItem(who, h)
	if err != nil {
		return rules.Outcome{}, err
	}
	half := c.Items[i].Ref.Count / 2
	if r.tables.Items[c.Items[i].Ref.ID].NotJoinable || half <= 0 {
		return refuseGear(gearCannotHalve)
	}
	if len(c.Items) >= maxCarriedItems { // the menu offers Halve only with room
		return rules.Outcome{}, rules.ErrNotNow
	}
	n := c.Items[i]
	n.Handle, n.Slot, n.Readied, n.Ref.Count = rules.NoItem, nextSlot(c), false, half
	c.Items[i].Ref.Count -= half
	c.Items = slices.Insert(c.Items, i+1, n)
	r.handOut(c)
	r.RecalcStats(c)
	return rules.Outcome{}, nil
}

// joinStack is JoinItem on c's item.
func (r *Rules) joinStack(who rules.CharID, h rules.ItemHandle) (rules.Outcome, error) {
	c, i, err := r.gearItem(who, h)
	if err != nil {
		return rules.Outcome{}, err
	}
	r.joinItem(c, i)
	r.RecalcStats(c)
	return rules.Outcome{}, nil
}

// giveItem is ItemTradeMenu after the pick: a readied item is refused (CheckItemAvail), a receiver already at
// maxCarriedItems is Overloaded (CannotCarryItem counts only the items, even for one that would join), else the
// item's copy goes to the end of the receiver's list, leaves the giver and joins the receiver's kind (not when the
// receiver is the giver).
func (r *Rules) giveItem(a rules.GiveItem) (rules.Outcome, error) {
	from, i, err := r.gearItem(a.From, a.Item)
	if err != nil {
		return rules.Outcome{}, err
	}
	to := r.chars[a.To]
	if to == nil {
		return rules.Outcome{}, errNoMember
	}
	if from.Items[i].Readied {
		return refuseGear(gearUnready)
	}
	if len(to.Items) >= maxCarriedItems {
		return refuseGear(gearOverloaded)
	}
	n := from.Items[i]
	n.Handle, n.Slot = rules.NoItem, nextSlot(to)
	to.Items = append(to.Items, n)
	r.handOut(to)
	from.Items = slices.Delete(from.Items, i, i+1) // after the copy: giving to oneself moves the item to the end
	if a.To != a.From {
		r.joinItem(to, len(to.Items)-1)
	}
	r.RecalcStats(from)
	r.RecalcStats(to)
	return rules.Outcome{}, nil
}

// nextSlot is the next free slot id on c's list.
func nextSlot(c *Character) rules.ItemSlot {
	slot := rules.ItemSlot(1)
	for _, x := range c.Items {
		slot = max(slot, x.Slot+1)
	}
	return slot
}
