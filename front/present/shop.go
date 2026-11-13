package present

import (
	"fmt"
	"slices"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/ui"
)

// The shop (play.Trading): Browse is the inventory with the goods; the
// presenter asks before an Id.
func (p *Presenter) shopItems() (string, []ui.MenuRow) {
	c, s := &p.camp, p.state()
	sh, purse := shopOf(s), play.CommonOf(s).Purse
	have := purse[Coin]
	unknown := p.eng.r.derived().unknown
	party := partyOf(s)
	switch c.screen {
	case shopID:
		if len(unknown) == 0 {
			break
		}
		var it []ui.MenuRow
		for i, u := range unknown {
			it = append(it, ui.MenuRow{Text: party[u.Member].Name + ": " + itemOf(party, u).Name, Key: 0, Do: func() { c.pick = i; c.to(shopDeal, 1)() }})
		}
		return withGold("Identify for "+gold(sh.IdentifyPrice), have), append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
	case shopDeal:
		if c.pick >= len(unknown) {
			c.screen = shopID
			return p.shopItems()
		}
		u := unknown[c.pick]
		return withGold(fmt.Sprintf("Identify %s for %s. Is it a deal?", itemOf(party, u).Name, gold(sh.IdentifyPrice)), have), []ui.MenuRow{
			{Text: "Yes", Key: 'Y', Do: func() { p.call(func(e *classicGame) { e.Identify(u.Member, u.Item) }); c.back(shopID)() }},
			{Text: "No", Key: 'N', Do: c.back(shopID)}}
	case appraise:
		return p.appraiseItems(purse, sh.Offers, c.back(campMain))
	}
	it := []ui.MenuRow{{Text: "Buy", Key: 'B', Do: func() { p.send(front.ActInventory, -1) }},
		{Text: "View", Key: 'V', Do: func() { p.send(front.ActMember, -1) }}} // look-only: the selected member's sheet
	it = append(it, p.moneyRows(allowedOf(s))...)
	if appraisable(purse, sh.Offers) {
		it = append(it, ui.MenuRow{Text: "Appraise", Key: 'A', Do: c.to(appraise, 0)})
	}
	if len(unknown) > 0 {
		it = append(it, ui.MenuRow{Text: "Identify", Key: 'I', Do: c.to(shopID, 0)})
	}
	title := "Shop"
	if pooled := sh.Pooled[Coin]; pooled > 0 {
		title = withGold(title, have) + " Pooled " + price(front.Money{Gold: pooled})
	}
	return title, append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }})
}

// itemOf is the party's item that r names (nil: none).
func itemOf(party []game.Member, r game.ItemRef) *game.Item {
	if r.Member < 0 || r.Member >= len(party) {
		return nil
	}
	if i := slices.IndexFunc(party[r.Member].Items, func(it game.Item) bool { return it.Handle == r.Item }); i >= 0 {
		return &party[r.Member].Items[i]
	}
	return nil
}

// unknownOf are the party's items the shop's Id would tell something about.
func unknownOf(s play.State) []game.ItemRef {
	sh := shopOf(s)
	if sh == nil {
		return nil
	}
	party := partyOf(s)
	return slices.DeleteFunc(slices.Clone(sh.Unknown), func(r game.ItemRef) bool { return itemOf(party, r) == nil })
}

// serviceAct takes the decisions of the shop, the temple and the training
// hall (and the party menu's Train, which opens the hall) as their rule
// calls: false leaves a to the engine.
func (p *Presenter) serviceAct(a front.Action) bool {
	v := p.shownView
	switch {
	case p.training || p.hallUp(): // the hall up, from Train or the game's own (ModeTraining)
		switch a.Kind {
		case front.ActClose:
			p.call(func(e *classicGame) { e.CloseTraining(); p.training = p.hallUp() })
		case front.ActLead: // another member: the game's training opens for them (Close would leave a hall the game opened)
			p.call(func(e *classicGame) { e.OpenTraining(a.N); p.training = p.hallUp() })
		default:
			return false
		}
	case a.Kind == front.ActTraining && v.MainMenu != nil:
		p.call(func(e *classicGame) { e.OpenTraining(a.N); p.training = p.hallUp() })
	case v.Shop != nil && a.Kind == front.ActInventory:
		p.call(func(e *classicGame) { e.Browse(a.N) })
	case (v.Temple != nil || v.Shop != nil) && a.Kind == front.ActMember:
		p.call(func(e *classicGame) { e.ViewMember(a.N) })
	case (v.Shop != nil || v.Temple != nil) && a.Kind == front.ActClose:
		p.call(func(e *classicGame) { e.Leave() })
	default:
		return false
	}
	return true
}

// hallUp: the game has the training hall open.
func (p *Presenter) hallUp() bool { return trainingOf(p.state()) != nil }

// ownTraining puts the training hall up in v, over the party menu.
func (p *Presenter) ownTraining(v front.View) front.View {
	if p.training && p.eng != nil {
		p.training = p.hallUp() // a game loaded closes it
	}
	if p.training {
		v.MainMenu, v.Hall, v.Logo, v.Powers = nil, false, nil, nil
	}
	return v
}

// moneyRows are the money acts allowed offers: Take, Pool, Share.
func (p *Presenter) moneyRows(allowed map[string]bool) (it []ui.MenuRow) {
	for _, o := range []struct {
		on   bool
		text string
		act  front.ActKind
	}{{allowed["Take"], "Take", front.ActTakeMoney}, {allowed["PoolCoins"], "Pool", front.ActLootPool}, {allowed["LootShare"], "Share", front.ActLootShare}} {
		if o.on {
			it = append(it, ui.MenuRow{Text: o.text, Key: o.text[0], Do: func() { p.send(o.act, 0) }})
		}
	}
	return it
}
