package ui

import (
	"fmt"
	"slices"

	"gobox/game"
)

// The temple (game.Trading): the presenter asks; Cure comes agreed.
func (p *Presenter) templeItems() (string, []MenuRow) {
	c, s := &p.camp, p.state()
	services, offers := templeOf(s)
	party, sel, purse := partyOf(s), p.selected(), game.CommonOf(s).Purse
	name := memberName(party, sel)
	var cure game.Service
	if c.pick >= 0 && c.pick < len(services) {
		cure = services[c.pick]
	}
	switch c.screen {
	case templeHeal:
		var it []MenuRow
		for i, cu := range services {
			cost := "" // a free service (Silver Blades' temple) shows no price, not 0g
			if cu.Price != 0 {
				cost = price(Money{Gold: cu.Price})
			}
			row := MenuRow{Text: fmt.Sprintf("%-20s %6s", cu.Name, cost), NoKey: true} // no digit: the prices stay in one column, the greyed rows' too
			if serviceNeeded(cu, party, sel) {                                         // else greyed out (Felix's ruling: only what would help the member): no key, click or button picks it
				row.Do = func() {
					c.pick = i
					c.to(templePay, 1)()
				}
			}
			it = append(it, row)
		}
		return withGold(name+", how can we help you?", purse[Coin]), append(it, MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
	case templePay:
		return withGold(fmt.Sprintf("%s will only cost %s. Pay for cure?", cure.Name, gold(cure.Price)), purse[Coin]), []MenuRow{
			{Text: "Yes", Key: 'Y', Do: func() {
				pick := c.pick
				p.call(func(e *classicGame) { e.do(game.Heal{Member: sel, Service: pick}) }) // the member by index, as Allowed lists it
				c.back(templeHeal)()
			}},
			{Text: "No", Key: 'N', Do: c.back(templeHeal)}}
	case appraise:
		return p.appraiseItems(purse, offers, c.back(campMain))
	}
	it := []MenuRow{{Text: "Heal", Key: 'H', Do: c.to(templeHeal, 0)}, {Text: "View", Key: 'V', Do: func() { p.send(ActMember, -1) }}}
	it = append(it, p.moneyRows(game.CommonOf(s).Allowed)...)
	if appraisable(purse, offers) {
		it = append(it, MenuRow{Text: "Appraise", Key: 'A', Do: c.to(appraise, 0)})
	}
	return "Temple", append(it, MenuRow{Text: "Exit", Key: 'E', Do: func() { p.send(ActClose, 0) }})
}

// serviceNeeded: the service would help the selected member (sel).
func serviceNeeded(sv game.Service, party []game.Member, sel int) bool {
	return sel >= 0 && sel < len(party) && slices.Contains(sv.Needs, sel)
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
