package present

import (
	"fmt"
	"gobox/front"
	"gobox/ui"
)

// The shop (View.Shop): Browse is the inventory with the goods; the
// presenter asks before an Id.
func (p *Presenter) shopItems(v front.View) (string, []ui.MenuRow) {
	c, s := &p.camp, v.Shop
	switch c.screen {
	case shopID:
		if len(s.Unknown) == 0 {
			break
		}
		var it []ui.MenuRow
		for i, u := range s.Unknown {
			it = append(it, ui.MenuRow{Text: u.Owner + ": " + u.Name, Key: 0, Do: func() { c.pick = i; c.to(shopDeal, 1)() }})
		}
		return withGold("Identify for "+gold(s.IDPrice), s.Gold), append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
	case shopDeal:
		if c.pick >= len(s.Unknown) {
			c.screen = shopID
			return p.shopItems(v)
		}
		u := s.Unknown[c.pick]
		return withGold(fmt.Sprintf("Identify %s for %s. Is it a deal?", u.Name, gold(s.IDPrice)), s.Gold), []ui.MenuRow{
			{Text: "Yes", Key: 'Y', Do: func() { p.call(func(e *classicGame) { e.Identify(u.Member, u.Item) }); c.back(shopID)() }},
			{Text: "No", Key: 'N', Do: c.back(shopID)}}
	case appraise:
		return p.appraiseItems(s.Appraise, c.back(campMain))
	}
	it := []ui.MenuRow{{Text: "Buy", Key: 'B', Do: func() { p.send(front.ActInventory, -1) }},
		{Text: "View", Key: 'V', Do: func() { p.send(front.ActMember, -1) }}} // look-only: the selected member's sheet
	it = append(it, p.moneyRows(front.Money3{Take: s.Take, Pool: s.Pool, Share: s.Share})...)
	if appraisable(s.Appraise) {
		it = append(it, ui.MenuRow{Text: "Appraise", Key: 'A', Do: c.to(appraise, 0)})
	}
	if len(s.Unknown) > 0 {
		it = append(it, ui.MenuRow{Text: "Identify", Key: 'I', Do: c.to(shopID, 0)})
	}
	title := "Shop"
	if s.Pooled > 0 {
		title = withGold(title, s.Gold) + " Pooled " + price(front.Money{Gold: s.Pooled})
	}
	return title, append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }})
}

// serviceAct takes the decisions of the shop, the temple and the training
// hall (and the party menu's Train, which opens the hall) as their rule
// calls: false leaves a to the engine.
func (p *Presenter) serviceAct(a front.Action) bool {
	v := p.shownView
	switch {
	case p.training || v.Training != nil: // the hall up, from Train or the game's own (ModeTraining)
		switch a.Kind {
		case front.ActClose:
			p.call(func(e *classicGame) { e.CloseTraining(); p.training = e.TrainingView() != nil })
		case front.ActLead: // another member: the game's training opens for them (Close would leave a hall the game opened)
			p.call(func(e *classicGame) { e.OpenTraining(a.N); p.training = e.TrainingView() != nil })
		default:
			return false
		}
	case a.Kind == front.ActTraining && p.mainMenuUp():
		p.call(func(e *classicGame) { e.OpenTraining(a.N); p.training = e.TrainingView() != nil })
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

// ownTraining puts the training hall up in v, over the party menu.
func (p *Presenter) ownTraining(v front.View) front.View {
	if p.training && p.eng != nil {
		v.Training = p.eng.TrainingView()
		p.training = v.Training != nil // a game loaded closes it
	}
	if p.training {
		v.MainMenu, v.Hall, v.Logo, v.Powers = nil, false, nil, nil
	}
	return v
}

// moneyRows are the money acts a place offers (m): Take, Pool, Share.
func (p *Presenter) moneyRows(m front.Money3) (it []ui.MenuRow) {
	for _, o := range []struct {
		on   bool
		text string
		act  front.ActKind
	}{{m.Take, "Take", front.ActTakeMoney}, {m.Pool, "Pool", front.ActLootPool}, {m.Share, "Share", front.ActLootShare}} {
		if o.on {
			it = append(it, ui.MenuRow{Text: o.text, Key: o.text[0], Do: func() { p.send(o.act, 0) }})
		}
	}
	return it
}
