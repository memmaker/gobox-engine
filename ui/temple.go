package ui

import (
	"fmt"
	"slices"

	"gobox/game"
)

// The temple (game.Trading): the services for the whole party, each
// offered (green) when it would help anyone; picked, the members it would
// help are asked for one after the other (AskForMembers), each paid and
// cured on the pick.
func (p *Presenter) templeItems() (string, []MenuRow) {
	c, s := &p.camp, p.state()
	services, offers := templeOf(s)
	purse := game.CommonOf(s).Purse
	switch c.screen {
	case templeHeal:
		var it []MenuRow
		for i, cu := range services {
			cost := "" // a free service (Silver Blades' temple) shows no price, not 0g
			if cu.Price != 0 {
				cost = price(Money{Gold: cu.Price})
			}
			row := MenuRow{Text: fmt.Sprintf("%-20s %6s", cu.Name, cost), NoKey: true} // no digit: the prices stay in one column, the greyed rows' too
			if len(cu.Needs) > 0 {                                                     // else greyed out: it would help no one, no key, click or button picks it
				row.Do = func() { p.templeAsk(i, cu.Name) }
			}
			it = append(it, row)
		}
		return withGold("How can we help you?", purse[Coin]), append(it, MenuRow{Text: "Exit", Key: 'E', Do: c.back(campMain)})
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

// templeAsk asks whom service i (name) is for: the members it would help
// as the state has them now, each picked paid and cured.
func (p *Presenter) templeAsk(i int, name string) {
	AskForMembers(p, name+" for whom?", func(n int) bool {
		services, _ := templeOf(p.state())
		return i < len(services) && slices.Contains(services[i].Needs, n)
	}, func(n int) {
		p.call(func(e *classicGame) { e.do(game.Heal{Member: n, Service: i}) }) // the member by index, as Needs lists it
	})
}

// templeFor is the temple service highlighted on its list (false: none is).
func (p *Presenter) templeFor() (game.Service, bool) {
	if p.q != nil || !p.templeUp() || p.camp.screen != templeHeal {
		return game.Service{}, false
	}
	services, _ := templeOf(p.state())
	if i := p.camp.sel; i >= 0 && i < len(services) {
		return services[i], true
	}
	return game.Service{}, false
}

// appraisable: the party has gems or jewelry to appraise (or one is being
// valued): else Appraise is not offered.
func appraisable(purse map[string]int, offers []game.Offer) bool {
	return purse[game.Gems]+purse[game.Jewelry] > 0 || len(offers) > 0 && offers[0].Value > 0
}
