package rulesxxvc

import (
	"fmt"

	"gobox/rules"
)

// takeItem is postcbt.pas LootTakeItemMenu's take: item i of the pile goes
// to who (Shop.GiveItemCopyToCreature) and, when it fits, leaves the pile
// (Took); over the item limit it stays (OverLoaded).
func (r *Rules) takeItem(s rules.Looting, a rules.TakeItem) (rules.Outcome, error) {
	c := r.chars[a.Who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("TakeItem: no character %d", a.Who)
	}
	if a.Item < 0 || int(a.Item) >= len(s.Items) {
		return rules.Outcome{}, fmt.Errorf("TakeItem: no item %d in the pile", a.Item)
	}
	o, ok := r.giveCopy(a.Who, c, s.Items[a.Item])
	if ok {
		o.Effects = append(o.Effects, rules.Took{Who: a.Who, Item: a.Item})
	}
	return o, nil
}

// takeMoney is money.pas TakeMoneyMenu/TakeMoney: the credits asked for, up
// to the pile's (InputNumberMax caps the number), go to who.
func (r *Rules) takeMoney(s rules.Looting, a rules.TakeMoney) (rules.Outcome, error) {
	c := r.chars[a.Who]
	if c == nil {
		return rules.Outcome{}, fmt.Errorf("TakeMoney: no character %d", a.Who)
	}
	if a.Amount < 0 || a.Amount > s.Money {
		return rules.Outcome{}, fmt.Errorf("TakeMoney: %d of %d credits", a.Amount, s.Money)
	}
	if r.options.SharedPurse {
		r.purse += a.Amount
	} else {
		c.Coins += a.Amount
	}
	return rules.Outcome{}, nil
}
