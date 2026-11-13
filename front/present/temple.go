package present

import (
	"fmt"
	"slices"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/ui"
)

// The temple (play.Trading): the presenter asks; Cure comes agreed.
func (p *Presenter) templeItems() (string, []ui.MenuRow) {
	c, s := &p.camp, p.state()
	services, offers := templeOf(s)
	party, sel, purse := partyOf(s), p.selected(), play.CommonOf(s).Purse
	name := memberName(party, sel)
	var cure game.Service
	if c.pick >= 0 && c.pick < len(services) {
		cure = services[c.pick]
	}
	switch c.screen {
	case templeHeal:
		var it []ui.MenuRow
		for i, cu := range services {
			cost := "" // a free service (Silver Blades' temple) shows no price, not 0g
			if cu.Price != 0 {
				cost = price(front.Money{Gold: cu.Price})
			}
			needed := serviceNeeded(cu, party, sel)
			row := ui.MenuRow{Text: fmt.Sprintf("%-20s %6s", cu.Name, cost), Key: 0, Do: func() {
				c.pick = i
				if needed {
					c.to(templePay, 1)()
				} else {
					c.to(templeNot, 1)()
				}
			}}
			if !needed && cu.Only { // greyed out: no key, click or button picks it
				row.Do = nil
			}
			it = append(it, row)
		}
		return withGold(name+", how can we help you?", purse[Coin]), append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
	case templeNot:
		not := cure.Not[sel]
		if cure.Declines { // the temple only says why
			return name + " " + not, []ui.MenuRow{{Text: "OK", Key: 'O', Do: c.back(templeHeal)}}
		}
		return name + " " + not + " Cast cure anyway?", []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: c.to(templePay, 1)}, {Text: "No", Key: 'N', Do: c.back(templeHeal)}}
	case templePay:
		return withGold(fmt.Sprintf("%s will only cost %s. Pay for cure?", cure.Name, gold(cure.Price)), purse[Coin]), []ui.MenuRow{
			{Text: "Yes", Key: 'Y', Do: func() { pick := c.pick; p.call(func(e *classicGame) { e.Cure(-1, pick) }); c.back(templeHeal)() }},
			{Text: "No", Key: 'N', Do: c.back(templeHeal)}}
	case appraise:
		return p.appraiseItems(purse, offers, c.back(campMain))
	}
	it := []ui.MenuRow{{Text: "Heal", Key: 'H', Do: c.to(templeHeal, 0)}, {Text: "View", Key: 'V', Do: func() { p.send(front.ActMember, -1) }}}
	it = append(it, p.moneyRows(play.CommonOf(s).Allowed)...)
	if appraisable(purse, offers) {
		it = append(it, ui.MenuRow{Text: "Appraise", Key: 'A', Do: c.to(appraise, 0)})
	}
	return "Temple", append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(front.ActClose, 0) }})
}

// serviceNeeded: the service would help the selected member (sel; no one
// selected: all).
func serviceNeeded(sv game.Service, party []game.Member, sel int) bool {
	return sel < 0 || sel >= len(party) || slices.Contains(sv.Needs, sel)
}

// memberName is party[sel]'s name ("": none selected).
func memberName(party []game.Member, sel int) string {
	if sel < 0 || sel >= len(party) {
		return ""
	}
	return party[sel].Name
}

// appraisable: the party has gems or jewelry to appraise (or one is being
// valued): else Appraise is not offered.
func appraisable(purse map[string]int, offers []game.Offer) bool {
	return purse[game.Gems]+purse[game.Jewelry] > 0 || len(offers) > 0 && offers[0].Value > 0
}
